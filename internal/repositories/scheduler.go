package repositories

import (
	"context"
	"fmt"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/engine/notification/core"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/conversion"
)

type schedulerRepo struct {
	queries *sqlc.Queries
}

type SchedulerRepo interface {
	GetDueScheduledNotifications(ctx context.Context, limit int) ([]*core.NotificationLog, error)
	GetDueRetryNotifications(ctx context.Context, limit int) ([]*core.NotificationLog, error)
	UpdateNotificationLog(ctx context.Context, params core.UpdateNotificationLogParams) error
	UpdateNotificationStatus(ctx context.Context, id string, status string) error
}

func NewSchedulerRepo(queries *sqlc.Queries) *schedulerRepo {
	return &schedulerRepo{queries: queries}
}

func (r *schedulerRepo) GetDueScheduledNotifications(ctx context.Context, limit int) ([]*core.NotificationLog, error) {
	rows, err := r.queries.GetDueScheduledNotifications(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	return r.mapLogs(rows), nil
}

func (r *schedulerRepo) GetDueRetryNotifications(ctx context.Context, limit int) ([]*core.NotificationLog, error) {
	rows, err := r.queries.GetDueRetryNotifications(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	return r.mapLogs(rows), nil
}

func (r *schedulerRepo) UpdateNotificationLog(ctx context.Context, params core.UpdateNotificationLogParams) error {
	rendered, err := conversion.JSONBFromMap(params.RenderedContent)
	if err != nil {
		return fmt.Errorf("failed to marshal rendered content: %w", err)
	}

	_, err = r.queries.UpdateNotificationLog(ctx, sqlc.UpdateNotificationLogParams{
		ID:              utils.MustStringToUUID(params.ID),
		Status:          params.Status,
		RenderedContent: rendered,
		AttemptCount:    int32(params.AttemptCount),
		SentAt:          conversion.TimestampFromPtr(params.SentAt),
		NextRetryAt:     conversion.TimestampFromPtr(params.NextRetryAt),
		ErrorMessage:    conversion.TextFromPtr(params.ErrorMessage),
	})
	return err
}

func (r *schedulerRepo) UpdateNotificationStatus(ctx context.Context, id string, status string) error {
	_, err := r.queries.UpdateNotificationStatus(ctx, sqlc.UpdateNotificationStatusParams{
		ID:     utils.MustStringToUUID(id),
		Status: status,
	})
	return err
}

func (r *schedulerRepo) mapLogs(rows []sqlc.NotificationLog) []*core.NotificationLog {
	logs := make([]*core.NotificationLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, MapToCoreLog(row))
	}
	return logs
}