package services

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/jackc/pgx/v5/pgtype"
)

type UpsertOAuthInput struct {
	Email      string
	FullName   string
	AvatarURL  string
	Provider   string
	ProviderID string
}

type CreateUser struct {
	Email        string
	PasswordHash string
	FullName     string
}

type UserService interface {
	UpsertOAuthUser(ctx context.Context, input UpsertOAuthInput) (*sqlc.User, error)
	FindUserByID(ctx context.Context, id pgtype.UUID) (*sqlc.User, error)
	FindUserByEmail(ctx context.Context, email string) (*sqlc.User, error)
	FindUserByProviderID(ctx context.Context, provider, providerID string) (*sqlc.User, error)
	CreateUser(ctx context.Context, params CreateUser) (*sqlc.User, error)
    GetFullUserDetails(ctx context.Context,userID pgtype.UUID)(*sqlc.GetUserWithWorkspaceRow,error)
}

type userService struct {
	repo repositories.UserRepository
}

func ( u *userService) GetFullUserDetails(ctx context.Context,userID pgtype.UUID)(*sqlc.GetUserWithWorkspaceRow,error){

  row, err := u.repo.GetUserWithWorkspace(ctx, userID)
    if err != nil {
        return nil, err
    }
    return &row, nil

}

func (u *userService) CreateUser(ctx context.Context, params CreateUser) (*sqlc.User, error) {

    args:=sqlc.CreateUserParams{
        Email: params.Email,
        FullName: params.FullName,
        PasswordHash: pgtype.Text{
            String: params.PasswordHash,
            Valid: params.PasswordHash!="",
        },
    }


  user,err:=  u.repo.CreateUser(ctx,args)

  return &user,err

}

func (u *userService) UpsertOAuthUser(ctx context.Context, input UpsertOAuthInput) (*sqlc.User, error) {
	user, err := u.repo.UpsertOAuthUser(ctx, sqlc.UpsertOAuthUserParams{
		Email:    input.Email,
		FullName: input.FullName,
		AvatarUrl: pgtype.Text{
			String: input.AvatarURL,
			Valid:  input.AvatarURL != "",
		},
		ProviderID: pgtype.Text{
			String: input.ProviderID,
			Valid:  input.ProviderID != "",
		},
		AuthProvider: input.Provider,
	})
	if err != nil {
		return nil, fmt.Errorf("upserting oauth user: %w", err)
	}
	return &user, nil
}

func (u *userService) FindUserByID(ctx context.Context, id pgtype.UUID) (*sqlc.User, error) {
	user, err := u.repo.FindUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %w", err)
	}
	return &user, nil
}
func (u *userService) FindUserByProviderID(ctx context.Context, provider, providerID string) (*sqlc.User, error) {

	params := sqlc.FindUserByProviderIDParams{
		AuthProvider: provider,
		ProviderID: pgtype.Text{
			String: providerID,
			Valid:  true,
		},
	}

	user, err := u.repo.FindUserByProviderID(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("finding user by provider id: %w", err)
	}
	return &user, nil
}
func (u *userService) FindUserByEmail(ctx context.Context, email string) (*sqlc.User, error) {
	user, err := u.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return &user, nil
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}
