package repositories

import (
	"context"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/models"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/internal/utils/helpers"
)

type SubscriberRepo interface {
	UpsertSubscriber(ctx context.Context, params sqlc.UpsertUserContactInfoParams) (*models.Subscriber, error)
	UpsertPreference(ctx context.Context, params sqlc.UpsertUserPreferenceParams) (*models.UserPreference, error)
	GetSubscriberByExternalIDAndChannel(ctx context.Context, workspaceID, envID, externalUserID, channel string) (*models.Subscriber, error)
}

type subscriberRepo struct {
	db *sqlc.Queries
}

func NewSubscriberRepo(db *sqlc.Queries) SubscriberRepo {
	return &subscriberRepo{db: db}
}

func (r *subscriberRepo) UpsertSubscriber(ctx context.Context, params sqlc.UpsertUserContactInfoParams) (*models.Subscriber, error) {
	row, err := r.db.UpsertUserContactInfo(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToSubscriber(row), nil
}

func (r *subscriberRepo) UpsertPreference(ctx context.Context, params sqlc.UpsertUserPreferenceParams) (*models.UserPreference, error) {
	row, err := r.db.UpsertUserPreference(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapToUserPreference(row), nil
}

func (r *subscriberRepo) GetSubscriberByExternalIDAndChannel(ctx context.Context, workspaceID, envID, externalUserID, channel string) (*models.Subscriber, error) {
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

func mapToSubscriber(row sqlc.UserInfo) *models.Subscriber {
	return &models.Subscriber{
		ID:             utils.UUIDToString(row.ID),
		WorkspaceID:    utils.UUIDToString(row.WorkspaceID),
		EnvironmentID:  utils.UUIDToString(row.EnvironmentID),
		ExternalUserID: row.ExternalUserID,
		Channel:        row.Channel,
		ContactValue:   row.ContactValue,
		IsVerified:     row.Verified.Bool,
		Metadata:       nil, 
		CreatedAt:      helpers.ToTime(row.CreatedAt),
		UpdatedAt:      helpers.ToTime(row.UpdatedAt),
	}
}

func mapToUserPreference(row sqlc.UserPreference) *models.UserPreference {
	return &models.UserPreference{
		ID:           utils.UUIDToString(row.ID),
		SubscriberID: utils.UUIDToString(row.SubscriberID),
		Channel:      row.Channel,
		EventType:    row.EventType.String,
		IsEnabled:    row.IsEnabled,
		CreatedAt:    helpers.ToTime(row.CreatedAt),
		UpdatedAt:    helpers.ToTime(row.UpdatedAt),
	}
}
