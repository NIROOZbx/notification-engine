package repositories

import (
	"context"
	"time"

	"github.com/NIROOZbx/notification-engine/db/sqlc"
	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/conversion"
)

type AnalyticsRepository interface {
	GetAggregateMetrics(ctx context.Context, workspaceID string, start, end time.Time) (*domain.AggregateMetrics, error)
	GetTimeSeriesData(ctx context.Context, workspaceID string, start, end time.Time, groupBy string) ([]domain.TimeSeriesData, error)
	GetProviderHealth(ctx context.Context, workspaceID string) ([]domain.ProviderHealth, error)
	GetLatencyTrend(ctx context.Context, workspaceID string) ([]int32, error)
	ListActivityLogs(ctx context.Context, workspaceID string, channel, status *string, limit, offset int32) ([]domain.ActivityLog, error)
	CountActivityLogs(ctx context.Context, workspaceID string, channel, status *string) (int64, error)
}

type analyticsRepository struct {
	queries *sqlc.Queries
}

func NewAnalyticsRepository(queries *sqlc.Queries) AnalyticsRepository {
	return &analyticsRepository{
		queries: queries,
	}
}

func (r *analyticsRepository) GetAggregateMetrics(ctx context.Context, workspaceID string, start, end time.Time) (*domain.AggregateMetrics, error) {
	rows, err := r.queries.GetAggregateMetrics(ctx, sqlc.GetAggregateMetricsParams{
		WorkspaceID: utils.MustStringToUUID(workspaceID),
		CreatedAt:   conversion.TimestampFromTime(start),
		CreatedAt_2: conversion.TimestampFromTime(end),
	})
	if err != nil {
		return nil, err
	}

	res := &domain.AggregateMetrics{
		ChannelCounts:  make(map[string]int64),
		ProviderCounts: make(map[string]int64),
	}

	for _, row := range rows {
		res.TotalSent += row.TotalCount
		res.TotalDelivered += row.DeliveredCount
		res.TotalFailed += row.FailedCount

		res.ChannelCounts[row.Channel] += row.TotalCount
		if row.Provider.Valid {
			res.ProviderCounts[row.Provider.String] += row.TotalCount
		}
	}

	return res, nil
}

func (r *analyticsRepository) GetTimeSeriesData(ctx context.Context, workspaceID string, start, end time.Time, groupBy string) ([]domain.TimeSeriesData, error) {
	rows, err := r.queries.GetTimeSeriesData(ctx, sqlc.GetTimeSeriesDataParams{
		WorkspaceID: utils.MustStringToUUID(workspaceID),
		CreatedAt:   conversion.TimestampFromTime(start),
		CreatedAt_2: conversion.TimestampFromTime(end),
		GroupBy:     groupBy,
	})
	if err != nil {
		return nil, err
	}

	res := make([]domain.TimeSeriesData, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.TimeSeriesData{
			Date:           conversion.TimeFromTimestampVal(row.TimeBucket),
			TotalSent:      row.TotalSent,
			TotalDelivered: row.TotalDelivered,
			TotalFailed:    row.TotalFailed,
		})
	}
	return res, nil
}

func (r *analyticsRepository) GetProviderHealth(ctx context.Context, workspaceID string) ([]domain.ProviderHealth, error) {
	rows, err := r.queries.GetProviderHealth(ctx, utils.MustStringToUUID(workspaceID))
	if err != nil {
		return nil, err
	}

	res := make([]domain.ProviderHealth, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.ProviderHealth{
			Provider:   row.Provider,
			AvgLatency: row.AvgLatency,
			LastSync:   conversion.TimeFromTimestampVal(row.LastSync),
		})
	}
	return res, nil
}

func (r *analyticsRepository) GetLatencyTrend(ctx context.Context, workspaceID string) ([]int32, error) {
	rows, err := r.queries.GetLatencyTrend(ctx, utils.MustStringToUUID(workspaceID))
	if err != nil {
		return nil, err
	}

	res := make([]int32, 0, len(rows))
	for _, row := range rows {
		res = append(res, row.Int32)
	}
	return res, nil
}

func (r *analyticsRepository) ListActivityLogs(ctx context.Context, workspaceID string, channel, status *string, limit, offset int32) ([]domain.ActivityLog, error) {
	rows, err := r.queries.ListActivityLogs(ctx, sqlc.ListActivityLogsParams{
		WorkspaceID: utils.MustStringToUUID(workspaceID),
		Channel:     conversion.ToNullString(channel),
		Status:      conversion.ToNullString(status),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, err
	}

	logs := make([]domain.ActivityLog, 0, len(rows))
	for _, row := range rows {
		triggerData, _ := conversion.MapFromJSONB(row.TriggerData)

		logs = append(logs, domain.ActivityLog{
			ID:                row.ID.String(),
			Channel:           row.Channel,
			DeliveryStatus:    row.DeliveryStatus.String,
			Recipient:         row.Recipient,
			Provider:          row.Provider.String,
			ProviderMessageID: row.ProviderMessageID.String,
			ProviderResponse:  row.ProviderResponse.String,
			ErrorMessage:      row.ErrorMessage.String,
			TemplateID:        row.TemplateID.String(),
			TemplateName:      row.TemplateName.String,
			ExternalUserID:    row.ExternalUserID,
			TriggerData:       triggerData,
			DurationMs:        row.DurationMs.Int32,
			AttemptCount:      row.AttemptCount,
			CreatedAt:         conversion.TimeFromTimestampVal(row.CreatedAt),
			SentAt:            conversion.PtrTimeFromTimestamp(row.SentAt),
			DeliveredAt:       conversion.PtrTimeFromTimestamp(row.DeliveredAt),
			FailedAt:          conversion.PtrTimeFromTimestamp(row.FailedAt),
		})
	}

	return logs, nil
}

func (r *analyticsRepository) CountActivityLogs(ctx context.Context, workspaceID string, channel, status *string) (int64, error) {
	count, err := r.queries.CountActivityLogs(ctx, sqlc.CountActivityLogsParams{
		WorkspaceID: utils.MustStringToUUID(workspaceID),
		Channel:     conversion.ToNullString(channel),
		Status:      conversion.ToNullString(status),
	})
	return count, err
}

