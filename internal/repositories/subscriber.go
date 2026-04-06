package repositories

import (
	"context"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
)

type SubscriberRepo interface {
	UpsertSubscriber(ctx context.Context, params sqlc.UpsertUserContactInfoParams) (sqlc.UserInfo, error)
	UpsertPreference(ctx context.Context, params sqlc.UpsertUserPreferenceParams) (sqlc.UserPreference, error)
}

type subscriberRepo struct {
	db *sqlc.Queries
}

func NewSubscriberRepo(db *sqlc.Queries) SubscriberRepo {
	return &subscriberRepo{db: db}
}

func (r *subscriberRepo) UpsertSubscriber(ctx context.Context, params sqlc.UpsertUserContactInfoParams) (sqlc.UserInfo, error) {
	return r.db.UpsertUserContactInfo(ctx, params)
}

func (r *subscriberRepo) UpsertPreference(ctx context.Context, params sqlc.UpsertUserPreferenceParams) (sqlc.UserPreference, error) {
	return r.db.UpsertUserPreference(ctx, params)
}