package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/NIROOZbx/notification-engine/config"
	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/dtos"
	"github.com/NIROOZbx/notification-engine/internal/session"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"

	"github.com/NIROOZbx/notification-engine/pkg/jwt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthService interface {
	HandleOAuthCallback(ctx context.Context, user *dtos.OAuthUserDetails) (*dtos.AuthResponse, *jwt.Pair, error)
	CompleteOnboarding(ctx context.Context, userID pgtype.UUID, workspaceName string) (*dtos.AuthResponse, *jwt.Pair, error)
	Register(ctx context.Context, req dtos.RegisterRequest) (*dtos.AuthResponse, *jwt.Pair, error)
	Login(ctx context.Context, req dtos.LoginRequest) (*dtos.AuthResponse, *jwt.Pair, error)
	Logout(ctx context.Context, userID pgtype.UUID, refreshToken string) error
}

type authService struct {
	authConfig   *config.AuthConfig
	userSvc      UserService
	workspaceSvc WorkspaceService
	store        session.Store
}

type sessionParams struct {
	userID      pgtype.UUID
	workspaceID pgtype.UUID
	role        string
	isOnboarded bool
}

func NewAuthService(cfg *config.AuthConfig,
	userSvc UserService,
	workspaceSvc WorkspaceService,
	store session.Store) AuthService {
	return &authService{
		userSvc:      userSvc,
		authConfig:   cfg,
		workspaceSvc: workspaceSvc,
		store:        store,
	}
}

func (a *authService) Register(ctx context.Context, req dtos.RegisterRequest) (*dtos.AuthResponse, *jwt.Pair, error) {

	hashedPassword, err := helpers.HashPassword(req.Password)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user, err := a.userSvc.CreateUser(ctx, CreateUser{
		Email:        req.Email,
		FullName:     req.Name,
		PasswordHash: hashedPassword,
	})

	if err != nil {
		if apperrors.IsUniqueViolation(err) {
			return nil, nil, apperrors.NewAlreadyExistsError("email")
		}
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}
	return a.buildAuthResult(ctx, user, nil, false, "")
}

func (a *authService) Login(ctx context.Context, req dtos.LoginRequest) (*dtos.AuthResponse, *jwt.Pair, error) {

	user, err := a.userSvc.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, apperrors.ErrUnauthorized
	}

	if !user.PasswordHash.Valid {
		return nil, nil, apperrors.ErrUnauthorized
	}

	err = helpers.ComparePassword(user.PasswordHash.String, req.Password)

	if err != nil {
		return nil, nil, apperrors.ErrUnauthorized
	}

	member, err := a.workspaceSvc.GetWorkspaceMemberByUserID(ctx, user.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, fmt.Errorf("checking membership: %w", err)
	}

	var workspace *dtos.WorkspaceResponse

	var role string

	if member != nil {
		wID, _ := utils.StringToUUID(member.WorkspaceID)
		workspace, err = a.workspaceSvc.GetByID(ctx, wID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch workspace: %w", err)
		}
		role = member.Role
	}

	return a.buildAuthResult(ctx, user, workspace, workspace != nil, role)

}

func (a *authService) Logout(ctx context.Context, userID pgtype.UUID, refreshToken string) error {
	if refreshToken != "" {
		claims, err := jwt.ParseRefreshToken(refreshToken, []byte(a.authConfig.RefreshTokenSecret))
		if err == nil {
			if err := a.store.DeleteRefreshToken(ctx, claims.TokenID); err != nil {
				log.Printf("failed to delete refresh token for user %s: %v", userID, err)
			}
		}
	}

	err := a.store.UpgradeTokenVersion(ctx, utils.UUIDToString(userID))
	if err != nil {
		log.Printf("failed to upgrade token version for user %s: %v", userID, err)
	}

	return nil
}

func (a *authService) HandleOAuthCallback(ctx context.Context, user *dtos.OAuthUserDetails) (*dtos.AuthResponse, *jwt.Pair, error) {

	params := UpsertOAuthInput{
		Email:      user.Email,
		FullName:   user.Name,
		AvatarURL:  user.AvatarURL,
		Provider:   user.Provider,
		ProviderID: user.UserID,
	}

	dbUser, err := a.userSvc.UpsertOAuthUser(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("upserting user: %w", err)
	}

	member, err := a.workspaceSvc.GetWorkspaceMemberByUserID(ctx, dbUser.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, fmt.Errorf("checking membership: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return a.buildAuthResult(ctx, dbUser, nil, false, "")
	}

	//-------------------------------------------------------------------------------//

	wID, _ := utils.StringToUUID(member.WorkspaceID)
	workspace, err := a.workspaceSvc.GetByID(ctx, wID)

	if err != nil {
		return nil, nil, fmt.Errorf("fetching workspace: %w", err)
	}

	return a.buildAuthResult(ctx, dbUser, workspace, true, member.Role)

}

func (a *authService) CompleteOnboarding(ctx context.Context, userID pgtype.UUID, workspaceName string) (*dtos.AuthResponse, *jwt.Pair, error) {

	wsp, err := a.workspaceSvc.GetOrCreate(ctx, userID, workspaceName)
	if err != nil {
		return nil, nil, fmt.Errorf("creating workspace: %w", err)
	}

	dbUser, err := a.userSvc.FindUserByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching user: %w", err)
	}

	return a.buildAuthResult(ctx, dbUser, wsp.Workspace, true, wsp.Role)
}

func (a *authService) generateSession(ctx context.Context, p sessionParams) (*jwt.Pair, error) {

	userIDStr := utils.UUIDToString(p.userID)
	if userIDStr == "" {
		return nil, fmt.Errorf("user_id is missing or invalid: %s", userIDStr)
	}
	wspIDStr := ""

	if p.isOnboarded {
		wspIDStr = utils.UUIDToString(p.workspaceID)
		if wspIDStr == "" {
			return nil, fmt.Errorf("onboarded user %s missing valid workspace_id", wspIDStr)
		}
	}

	version, err := a.store.GetTokenVersion(ctx, userIDStr)
	if err != nil {
		log.Printf("redis down, failing open for user %s: %v", userIDStr, err)
		version = 0
	}

	payload := &jwt.TokenPayload{
		Role:        p.role,
		UserID:      userIDStr,
		WorkspaceID: wspIDStr,
		Version:     version,
		IsOnboarded: p.isOnboarded,
	}

	jwtConfig := a.authConfig.ToJWTConfig()

	tokenPair, err := jwt.GenerateTokenPair(jwtConfig, *payload)
	if err != nil {
		return nil, fmt.Errorf("generating token pair: %w", err)
	}

	expiry := time.Duration(a.authConfig.RefreshExpiryHours) * time.Hour

	err = a.store.StoreRefreshToken(ctx, tokenPair.TokenID, userIDStr, expiry)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil

}

func mapToDTO(user *sqlc.User, wsp *dtos.WorkspaceResponse, role string) (*dtos.AuthResponse, error) {
	userID := utils.UUIDToString(user.ID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is missing or invalid: %s", userID)
	}

	avatarURL := ""
	if user.AvatarUrl.Valid {
		avatarURL = user.AvatarUrl.String
	}

	dto := &dtos.AuthResponse{
		User: dtos.UserDetails{
			UserID:       userID,
			Email:        user.Email,
			Name:         user.FullName,
			AvatarURL:    avatarURL,
			HasWorkspace: wsp != nil,
		},
	}

	if wsp != nil {
		if wsp.ID == "" {
			return nil, fmt.Errorf("missing valid workspace_id in fetched workspace")
		}

		dto.Workspace = &dtos.WorkSpaceDetails{
			WorkspaceID:   wsp.ID,
			WorkSpaceName: wsp.Name,
			Slug:          wsp.Slug,
			Role:          role,
			Environments:  wsp.Environments,
		}

	}

	return dto, nil

}

func (a *authService) buildAuthResult(ctx context.Context, user *sqlc.User, workspace *dtos.WorkspaceResponse, isOnboarded bool, role string) (*dtos.AuthResponse, *jwt.Pair, error) {
	if user == nil {
		return nil, nil, fmt.Errorf("developer error: user cannot be nil in buildAuthResult")
	}

	var workspaceID pgtype.UUID

	if workspace != nil {
		workspaceID, _ = utils.StringToUUID(workspace.ID)
	}
	p := sessionParams{
		userID:      user.ID,
		workspaceID: workspaceID,
		role:        role,
		isOnboarded: isOnboarded,
	}
	pair, err := a.generateSession(ctx, p)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate session: %w", err)
	}

	dto, err := mapToDTO(user, workspace, role)
	if err != nil {
		return nil, nil, fmt.Errorf("error in mapping dto: %w", err)
	}

	return dto, pair, nil

}
