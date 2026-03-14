package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/NIROOZbx/notification-engine/services/backend/config"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/db"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/dtos"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/session"
	userService "github.com/NIROOZbx/notification-engine/services/backend/internal/user/services"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/utils"
	workspaceSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/workspace/services"
	"github.com/NIROOZbx/notification-engine/services/pkg/jwt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthService interface {
	HandleOAuthCallback(ctx context.Context, user *dtos.OAuthUserDetails) (*dtos.AuthResponse, *jwt.Pair, error)
	CompleteOnboarding(ctx context.Context, userID pgtype.UUID, workspaceName string) (*dtos.AuthResponse, *jwt.Pair, error)
	Logout(ctx context.Context, userID string, refreshToken string) error 
}

type authService struct {
	authConfig   *config.AuthConfig
	userSvc      userService.UserService
	workspaceSvc workspaceSvc.WorkspaceService
	store        session.Store
}

type sessionParams struct {
    userID      pgtype.UUID
    workspaceID pgtype.UUID
    role        string
    isOnboarded bool
}

func NewAuthService(cfg *config.AuthConfig,
	userSvc userService.UserService,
	workspaceSvc workspaceSvc.WorkspaceService,
	store session.Store) AuthService {
	return &authService{
		userSvc:      userSvc,
		authConfig:   cfg,
		workspaceSvc: workspaceSvc,
		store:        store,
	}
}



func (a *authService) Logout(ctx context.Context, userID string, refreshToken string) error {
	if refreshToken != "" {
		claims, err := jwt.ParseRefreshToken(refreshToken, []byte(a.authConfig.RefreshTokenSecret))
		if err == nil {
			if err:= a.store.DeleteRefreshToken(ctx, claims.TokenID);err!=nil{
				log.Printf("failed to delete refresh token for user %s: %v", userID, err)
			}
		}
	}

	err := a.store.UpgradeTokenVersion(ctx, userID)
	if err != nil {
		log.Printf("failed to upgrade token version for user %s: %v", userID, err)
	}

	return nil
}

func (a *authService) HandleOAuthCallback(ctx context.Context, user *dtos.OAuthUserDetails) (*dtos.AuthResponse, *jwt.Pair, error) {

	params := userService.UpsertOAuthInput{
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

		p:=sessionParams{
			userID:  dbUser.ID,
			workspaceID: pgtype.UUID{},
			role: "",
			isOnboarded: false,
		}
		pair, err := a.generateSession(ctx,p)
		if err != nil {
			return nil, nil, err
		}

		dto, err := mapToDTO(dbUser, nil, "")
		if err != nil {
			return nil, nil, err
		}
		return dto, pair, nil
	}

	//-------------------------------------------------------------------------------//

	workspace, err := a.workspaceSvc.GetByID(ctx, member.WorkspaceID)

	if err != nil {
		return nil, nil, fmt.Errorf("fetching workspace: %w", err)
	}

	
		p:=sessionParams{
			userID:  dbUser.ID,
			workspaceID: member.WorkspaceID,
			role: member.Role,
			isOnboarded: true,
		}

	pair, err := a.generateSession(ctx,p)

	if err != nil {
		return nil, nil, fmt.Errorf("generating session: %w", err)
	}

	dto, err := mapToDTO(dbUser, workspace, member.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("error in mapping dto: %w", err)
	}

	return dto, pair, nil
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

	p:=sessionParams{
			userID: userID,
			workspaceID:  wsp.Workspace.ID,
			role: wsp.Role,
			isOnboarded: true,
		}

	pair, err := a.generateSession(ctx,p)
	if err != nil {
		return nil, nil, err
	}

	dto, err := mapToDTO(dbUser, wsp.Workspace, wsp.Role)
	if err != nil {
		return nil, nil, fmt.Errorf("error in mapping dto: %w", err)
	}

	return dto, pair, nil

}

func (a *authService) generateSession(ctx context.Context, p sessionParams) (*jwt.Pair, error) {

	userIDStr, err := utils.UUIDToString(p.userID)
	if err != nil {
		return nil, fmt.Errorf("converting user id: %w", err)
	}
	wspIDStr := ""

	if p.isOnboarded {
		wspIDStr, err = utils.UUIDToString(p.workspaceID)
		if err != nil {
			return nil, fmt.Errorf("converting workspace id: %w", err)
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

func mapToDTO(user *db.User, wsp *db.Workspace, role string) (*dtos.AuthResponse, error) {
	userID, err := utils.UUIDToString(user.ID)
	if err != nil {
		return nil, fmt.Errorf("converting user id: %w", err)
	}

	avatarURL := ""
	if user.AvatarUrl.Valid {
		avatarURL = user.AvatarUrl.String
	}

	dto := &dtos.AuthResponse{
		User: dtos.UserDetails{
			UserID:    userID,
			Email:     user.Email,
			Name:      user.FullName,
			AvatarURL: avatarURL,
		},
	}

	var workspaceID string

	if wsp != nil {
		workspaceID, err = utils.UUIDToString(wsp.ID)
		if err != nil {
			return nil, fmt.Errorf("converting workspace id: %w", err)
		}
		
		dto.Workspace = &dtos.WorkSpaceDetails{
			WorkspaceID:   workspaceID,
			WorkSpaceName: wsp.Name,
			Slug:          wsp.Slug,
			Role:          role,
		}

	}

	return dto, nil

}
