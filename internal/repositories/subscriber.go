package repositories

import (
	"context"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
	"github.com/NIROOZbx/notification-engine/pkg/serializer"
	"github.com/jackc/pgx/v5/pgtype"
)

type SubscriberRepo interface {
	UpsertSubscriber(ctx context.Context, params sqlc.UpsertUserContactInfoParams) (*domain.Subscriber, error)
	UpsertPreference(ctx context.Context, params sqlc.UpsertUserPreferenceParams) (*domain.UserPreference, error)
	GetSubscriberByExternalIDAndChannel(ctx context.Context, workspaceID, envID, externalUserID, channel string) (*domain.Subscriber, error)
	DeleteSubscriber(ctx context.Context, id, workspaceID pgtype.UUID) error
	ListSubscribers(ctx context.Context, workspaceID, environmentID pgtype.UUID, limit, offset int32) ([]*domain.Subscriber, error)
	CountSubscribers(ctx context.Context, workspaceID, environmentID pgtype.UUID) (int64, error)
	ListPreferences(ctx context.Context, workspaceID, environmentID pgtype.UUID, externalUserID string) ([]*domain.UserPreference, error)
}

type subscriberRepo struct {
	db *sqlc.Queries
}

func NewSubscriberRepo(db *sqlc.Queries) *subscriberRepo {
	return &subscriberRepo{db: db}
}

func (r *subscriberRepo) UpsertSubscriber(ctx context.Context, params sqlc.UpsertUserContactInfoParams) (*domain.Subscriber, error) {
	row, err := r.db.UpsertUserContactInfo(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToSubscriber(row), nil
}

func (r *subscriberRepo) UpsertPreference(ctx context.Context, params sqlc.UpsertUserPreferenceParams) (*domain.UserPreference, error) {
	row, err := r.db.UpsertUserPreference(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToUserPreference(row), nil
}

func (r *subscriberRepo) GetSubscriberByExternalIDAndChannel(ctx context.Context, workspaceID, envID, externalUserID, channel string) (*domain.Subscriber, error) {
	row, err := r.db.GetContactByExternalUserAndChannel(ctx, sqlc.GetContactByExternalUserAndChannelParams{
		WorkspaceID:    utils.MustStringToUUID(workspaceID),
		EnvironmentID:  utils.MustStringToUUID(envID),
		ExternalUserID: externalUserID,
		Channel:        channel,
	})
	if err != nil {
		return nil, err
	}

	return mapToSubscriber(row), nil
}

func (r *subscriberRepo) DeleteSubscriber(ctx context.Context, id, workspaceID pgtype.UUID) error {
	return r.db.DeleteUserContactInfo(ctx, sqlc.DeleteUserContactInfoParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
}

func (r *subscriberRepo) ListSubscribers(ctx context.Context, workspaceID, environmentID pgtype.UUID, limit, offset int32) ([]*domain.Subscriber, error) {
	rows, err := r.db.ListSubscribers(ctx, sqlc.ListSubscribersParams{
		WorkspaceID:   workspaceID,
		EnvironmentID: environmentID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}

	subscribers := make([]*domain.Subscriber, len(rows))
	for i, row := range rows {
		subscribers[i] = mapToSubscriber(row)
	}

	return subscribers, nil
}

func (r *subscriberRepo) CountSubscribers(ctx context.Context, workspaceID, environmentID pgtype.UUID) (int64, error) {
	return r.db.CountSubscribers(ctx, sqlc.CountSubscribersParams{
		WorkspaceID:   workspaceID,
		EnvironmentID: environmentID,
	})
}

func (r *subscriberRepo) ListPreferences(ctx context.Context, workspaceID, environmentID pgtype.UUID, externalUserID string) ([]*domain.UserPreference, error) {
	rows, err := r.db.ListUserPreferencesBySubscriber(ctx, sqlc.ListUserPreferencesBySubscriberParams{
		WorkspaceID:    workspaceID,
		EnvironmentID:  environmentID,
		ExternalUserID: externalUserID,
	})
	if err != nil {
		return nil, err
	}

	prefs := make([]*domain.UserPreference, len(rows))
	for i, row := range rows {
		prefs[i] = mapToUserPreference(row)
	}

	return prefs, nil
}




func mapToSubscriber(row sqlc.UserInfo) *domain.Subscriber {
	var metadata map[string]any
	if len(row.Metadata) > 0 {
		_ = serializer.Unmarshal(row.Metadata, &metadata)
	}

	return &domain.Subscriber{
		ID:             utils.UUIDToString(row.ID),
		WorkspaceID:    utils.UUIDToString(row.WorkspaceID),
		EnvironmentID:  utils.UUIDToString(row.EnvironmentID),
		ExternalUserID: row.ExternalUserID,
		Channel:        row.Channel,
		ContactValue:   row.ContactValue,
		IsVerified:     row.Verified.Bool,
		Metadata:       metadata,
		CreatedAt:      helpers.ToTime(row.CreatedAt),
		UpdatedAt:      helpers.ToTime(row.UpdatedAt),
	}
}

func mapToUserPreference(row sqlc.UserPreference) *domain.UserPreference {
	return &domain.UserPreference{
		ID:           utils.UUIDToString(row.ID),
		SubscriberID: utils.UUIDToString(row.SubscriberID),
		Channel:      row.Channel,
		EventType:    row.EventType.String,
		IsEnabled:    row.IsEnabled,
		CreatedAt:    helpers.ToTime(row.CreatedAt),
		UpdatedAt:    helpers.ToTime(row.UpdatedAt),
	}
}
