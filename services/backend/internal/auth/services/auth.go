package services

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/services/backend/config"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/session"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/db"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/dtos"
	userService "github.com/NIROOZbx/notification-engine/services/backend/internal/user/services"
	"github.com/NIROOZbx/notification-engine/services/backend/internal/utils"
	workspaceSvc "github.com/NIROOZbx/notification-engine/services/backend/internal/workspace/services"
	"github.com/NIROOZbx/notification-engine/services/pkg/jwt"
)

type AuthService interface {
	HandleOAuthCallback(ctx context.Context, user *dtos.OAuthUserDetails) (*dtos.UserDTO, *jwt.Pair, error)
}

type authService struct {
	authConfig *config.AuthConfig
	userSvc      userService.UserService
	workspaceSvc workspaceSvc.WorkspaceService
	store session.Store
}

func NewAuthService(repo *db.Queries, cfg *config.AuthConfig,userSvc userService.UserService) AuthService {
	return &authService{
		userSvc:      userSvc,
		authConfig: cfg,
	}
}
func (a *authService) HandleOAuthCallback(ctx context.Context, user *dtos.OAuthUserDetails) (*dtos.UserDTO, *jwt.Pair, error) {

	params := userService.UpsertOAuthInput{
		Email: user.Email,
		FullName: user.Name,
		AvatarURL: user.AvatarURL,
		Provider: user.Provider,
		ProviderID: user.UserID,
	}

	dbUser, err := a.userSvc.UpsertOAuthUser(ctx,params)
	if err != nil {
        return nil, nil, fmt.Errorf("upserting user: %w", err)
    }

	pair,err:=a.mintTokens(ctx,dbUser)

	if err!=nil{
		return nil,nil,fmt.Errorf("error in mintin tokens: %w", err)
	}

	

	userData,err:=a.MapToDTO(user)

	return userData, pair, nil
}


func (a *authService)mintTokens(ctx context.Context,user *db.User)(*jwt.Pair,error){

	userID, err := utils.UUIDToString(user.ID)
	if err != nil {
		return nil, fmt.Errorf("converting user id: %w", err)
	}

	version,err:=a.store.GetTokenVersion(ctx,userID)

	if err!=nil{
		return nil,err
	}

	payload := &jwt.TokenPayload{
		Role:        "",
		UserID:      userID,
		WorkspaceID: "",
		Version: version,
	}
	jwtConfig:=a.ToJWTConfig()

	tokenPair, err := jwt.GenerateTokenPair(jwtConfig, *payload)
	if err != nil {
		return nil, fmt.Errorf("generating token pair: %w", err)
	}

	return tokenPair,nil

}

func (a *authService) MapToDTO(user *db.User, wsp *db.Workspace, role string) (*dtos.UserDTO, error) {
	userID, err := utils.UUIDToString(user.ID)
	if err != nil {
		return nil, fmt.Errorf("converting user id: %w", err)
	}

	workspaceID, err := utils.UUIDToString(wsp.ID)
	if err != nil {
		return nil, fmt.Errorf("converting workspace id: %w", err)
	}

	avatarURL := ""
	if user.AvatarUrl.Valid {
		avatarURL = user.AvatarUrl.String
	}
	return &dtos.UserDTO{
		UserDetails: dtos.UserDetails{
			UserID:    userID,
			Email:     user.Email,
			Name:      user.FullName,
			AvatarURL: avatarURL,
		},
		WorkSpaceDetails: dtos.WorkSpaceDetails{
			WorkspaceID:   workspaceID,
			WorkSpaceName: wsp.Name,
			Slug:          wsp.Slug,
			Role:          role,
		},
	}, nil

}
func (a *authService) ToJWTConfig() jwt.Config {
    return jwt.Config{
        AccessTokenSecret:   a.authConfig.AccessTokenSecret,
        RefreshTokenSecret:  a.authConfig.RefreshTokenSecret,
        AccessExpiryMinutes: a.authConfig.AccessExpiryMinutes,
        RefreshExpiryHours:  a.authConfig.RefreshExpiryHours,
    }
}