package repositories

import (
	"context"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
)

type channelConfigRepo struct {
	queries *sqlc.Queries
}

type ChannelConfigRepo interface {
	Create(ctx context.Context, encryptedString string, params domain.CreateChannelConfigParams) (*domain.ChannelConfig, error)
	List(ctx context.Context, workspaceID pgtype.UUID) ([]*domain.ChannelConfig, error)
	GetChannelConfigByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.ChannelConfig, error)
	GetDefaultChannelConfig(ctx context.Context, workspaceID pgtype.UUID, channel string) (*domain.ChannelConfig, error)
	UpdateChannelConfig(ctx context.Context, encryptedString *string, params domain.UpdateChannelConfigParams) (*domain.ChannelConfig, error)
	SetChannelConfigDefault(ctx context.Context, id, workspaceID pgtype.UUID) error
	RemoveProviderOverride(ctx context.Context, id, workspaceID pgtype.UUID) error
	SoftDeleteChannelConfig(ctx context.Context, id, workspaceID pgtype.UUID) error
	DeleteChannelConfig(ctx context.Context, id, workspaceID pgtype.UUID) error
	CountActiveProvidersForChannel(ctx context.Context, workspaceID pgtype.UUID, channel string) (int64, error)
}

func NewChannelConfigRepo(queries *sqlc.Queries) *channelConfigRepo {
	return &channelConfigRepo{queries: queries}
}

func (c *channelConfigRepo) Create(ctx context.Context, encryptedString string, params domain.CreateChannelConfigParams) (*domain.ChannelConfig, error) {
	row, err := c.queries.CreateChannelConfig(ctx, sqlc.CreateChannelConfigParams{
		WorkspaceID: params.WorkspaceID,
		Channel:     params.Channel,
		Provider:    params.Provider,
		DisplayName: helpers.Text(params.DisplayName),
		Credentials: encryptedString,
		IsActive:    params.IsActive,
		IsDefault:   params.IsDefault,
	})

	if err != nil {
		return nil, apperrors.MapDBError(err)
	}

	return mapToChannelConfig(row), nil
}

func (c *channelConfigRepo) List(ctx context.Context, workspaceID pgtype.UUID) ([]*domain.ChannelConfig, error) {
	rows, err := c.queries.ListChannelConfigs(ctx, workspaceID)
	if err != nil {
		return nil, apperrors.MapDBError(err)
	}

	return mapToChannelConfigs(rows), nil
}

func (c *channelConfigRepo) GetChannelConfigByID(ctx context.Context, id, workspaceID pgtype.UUID) (*domain.ChannelConfig, error) {
	channel, err := c.queries.GetChannelConfigByID(ctx, sqlc.GetChannelConfigByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})

	if err != nil {
		return nil, apperrors.MapDBError(err)
	}

	return mapToChannelConfig(channel), nil
}

func (c *channelConfigRepo) GetDefaultChannelConfig(ctx context.Context, workspaceID pgtype.UUID, channel string) (*domain.ChannelConfig, error) {
	row, err := c.queries.GetDefaultChannelConfig(ctx, sqlc.GetDefaultChannelConfigParams{
		WorkspaceID: workspaceID,
		Channel:     channel,
	})
	if err != nil {
		return nil, apperrors.MapDBError(err)
	}
	return mapToChannelConfig(row), nil
}

func (c *channelConfigRepo) UpdateChannelConfig(ctx context.Context, encryptedString *string, params domain.UpdateChannelConfigParams) (*domain.ChannelConfig, error) {
	row, err := c.queries.UpdateChannelConfig(ctx, sqlc.UpdateChannelConfigParams{
		DisplayName: helpers.TextPtr(params.DisplayName),
		Credentials: helpers.TextPtr(encryptedString),
		IsActive:    helpers.BoolPtr(params.IsActive),
		WorkspaceID: params.WorkspaceID,
		ID:          params.ID,
	})
	if err != nil {
		return nil, apperrors.MapDBError(err)
	}
	return mapToChannelConfig(row), nil
}

func (c *channelConfigRepo) DeleteChannelConfig(ctx context.Context, id, workspaceID pgtype.UUID) error {

	return c.queries.SoftDeleteChannelConfig(ctx, sqlc.SoftDeleteChannelConfigParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
}

func (c *channelConfigRepo) SetChannelConfigDefault(ctx context.Context, id, workspaceID pgtype.UUID) error {

	current, err := c.GetChannelConfigByID(ctx, id, workspaceID)
	if err != nil {
		return err
	}

	err = c.queries.SetChannelConfigDefault(ctx, sqlc.SetChannelConfigDefaultParams{
		ID:          id,
		WorkspaceID: workspaceID,
		Channel:     current.Channel,
	})
	if err != nil {
		return apperrors.MapDBError(err)
	}
	return nil
}

func (c *channelConfigRepo) CountActiveProvidersForChannel(ctx context.Context, workspaceID pgtype.UUID, channel string) (int64, error) {
	return c.queries.CountActiveProvidersForChannel(ctx, sqlc.CountActiveProvidersForChannelParams{
		WorkspaceID: workspaceID,
		Channel:     channel,
	})
}

func (c *channelConfigRepo) RemoveProviderOverride(ctx context.Context, id, workspaceID pgtype.UUID) error {
	return c.queries.RemoveProviderOverride(ctx, sqlc.RemoveProviderOverrideParams{
		ChannelConfigID: id,
		WorkspaceID:     workspaceID,
	})
}

func (c *channelConfigRepo) SoftDeleteChannelConfig(ctx context.Context, id, workspaceID pgtype.UUID) error {
	return c.queries.SoftDeleteChannelConfig(ctx, sqlc.SoftDeleteChannelConfigParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
}

func mapToChannelConfig(row sqlc.ChannelConfig) *domain.ChannelConfig {
	return &domain.ChannelConfig{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		Channel:     row.Channel,
		Provider:    row.Provider,
		DisplayName: row.DisplayName.String,
		Encrypted:   row.Credentials,
		Credentials: nil,
		IsActive:    row.IsActive,
		IsDefault:   row.IsDefault,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func mapToChannelConfigs(rows []sqlc.ChannelConfig) []*domain.ChannelConfig {
	result := make([]*domain.ChannelConfig, len(rows))
	for i, row := range rows {
		result[i] = mapToChannelConfig(row)
	}
	return result
}
