package services

import (
	"context"
	"sort"
	"time"

	"github.com/NIROOZbx/notification-engine/internal/domain"
	"github.com/NIROOZbx/notification-engine/internal/repositories"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type AnalyticsService interface {
	GetFullAnalytics(ctx context.Context, workspaceID string, start, end time.Time, groupBy string) (*domain.AnalyticsResponse, error)
	GetActivityLogs(ctx context.Context, workspaceID string, channel, status *string, page, pageSize int32) (*domain.ActivityLogResponse, error)
}

type analyticsService struct {
	repo repositories.AnalyticsRepository
	log  zerolog.Logger
}

type buildResponseParams struct {
	currentMetrics  *domain.AggregateMetrics
	previousMetrics *domain.AggregateMetrics
	timeSeries      []domain.TimeSeriesData
	providerHealth  []domain.ProviderHealth
	latencyTrend    []int32
}

func NewAnalyticsService(repo repositories.AnalyticsRepository, log zerolog.Logger) *analyticsService {
	return &analyticsService{
		repo: repo,
		log:  log,
	}
}

func (s *analyticsService) GetFullAnalytics(ctx context.Context, workspaceID string, start, end time.Time, groupBy string) (*domain.AnalyticsResponse, error) {

	duration := end.Sub(start)
	prevStart := start.Add(-duration)
	prevEnd := start

	var (
		currentMetrics  *domain.AggregateMetrics
		previousMetrics *domain.AggregateMetrics
		timeSeries      []domain.TimeSeriesData
		providerHealth  []domain.ProviderHealth
		latencyTrend    []int32
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		currentMetrics, err = s.repo.GetAggregateMetrics(gctx, workspaceID, start, end)
		return err
	})

	g.Go(func() error {
		var err error
		previousMetrics, err = s.repo.GetAggregateMetrics(gctx, workspaceID, prevStart, prevEnd)
		return err
	})
	g.Go(func() error {
		var err error
		latencyTrend, err = s.repo.GetLatencyTrend(gctx, workspaceID)
		return err
	})
	g.Go(func() error {
		var err error

		providerHealth, err = s.repo.GetProviderHealth(gctx, workspaceID)
		return err
	})
	g.Go(func() error {
		var err error

		timeSeries, err = s.repo.GetTimeSeriesData(gctx, workspaceID, start, end, groupBy)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return s.buildResponse(buildResponseParams{
		currentMetrics:  currentMetrics,
		previousMetrics: previousMetrics,
		timeSeries:      timeSeries,
		providerHealth:  providerHealth,
		latencyTrend:    latencyTrend,
	}, groupBy)

}

func (s *analyticsService) buildResponse(params buildResponseParams, groupBy string) (*domain.AnalyticsResponse, error) {
	resp := &domain.AnalyticsResponse{}

	s.mapAggregates(resp, params)
	s.mapProviders(resp, params)
	s.mapTimeSeries(resp, params, groupBy)
	s.mapHealth(resp, params)

	return resp, nil
}

func (s *analyticsService) mapAggregates(resp *domain.AnalyticsResponse, params buildResponseParams) {
	curr := params.currentMetrics
	resp.Aggregate.TotalSent = curr.TotalSent
	resp.Aggregate.TotalDelivered = curr.TotalDelivered
	resp.Aggregate.TotalFailed = curr.TotalFailed
	resp.Aggregate.TotalBounced = curr.TotalBounced

	resp.Aggregate.Trends = buildTrends(curr, params.previousMetrics)
	resp.Aggregate.MostUsedChannel = findMostUsed(curr.ChannelCounts)
	resp.Aggregate.MostUsedProvider = findMostUsed(curr.ProviderCounts)

	resp.Channels = curr.ChannelCounts
}

func (s *analyticsService) mapProviders(resp *domain.AnalyticsResponse, params buildResponseParams) {
	resp.Providers = make([]domain.ProviderCount, 0, len(params.currentMetrics.ProviderCounts))
	for k, val := range params.currentMetrics.ProviderCounts {
		resp.Providers = append(resp.Providers, domain.ProviderCount{
			Name:  k,
			Count: val,
		})
	}

	sort.Slice(resp.Providers, func(i, j int) bool {
		return resp.Providers[i].Count > resp.Providers[j].Count
	})
}

func (s *analyticsService) mapTimeSeries(resp *domain.AnalyticsResponse, params buildResponseParams, groupBy string) {
	resp.TimeSeries = make([]domain.TimeSeriesDataDto, 0, len(params.timeSeries))
	for _, ts := range params.timeSeries {
		label := ts.Date.Format("02 Jan")
		switch groupBy {
		case "hour":
			label = ts.Date.Format("15:04")
		case "week":
			label = ts.Date.Format("02 Jan")

		case "month":
			label = ts.Date.Format("Jan 2006")
		}

		resp.TimeSeries = append(resp.TimeSeries, domain.TimeSeriesDataDto{
			Label:          label,
			SentCount:      ts.TotalSent,
			DeliveredCount: ts.TotalDelivered,
			FailedCount:    ts.TotalFailed,
		})
	}
}

func (s *analyticsService) mapHealth(resp *domain.AnalyticsResponse, params buildResponseParams) {
	resp.Health.ActiveProviders = params.providerHealth
	resp.Health.LatencyTrend = params.latencyTrend

	if len(params.providerHealth) > 0 {
		resp.Aggregate.MostRecentProvider = params.providerHealth[0].Provider

		var totalLatency int32
		for _, p := range params.providerHealth {
			totalLatency += p.AvgLatency
		}
		resp.Health.AverageLatencyMs = totalLatency / int32(len(params.providerHealth))
	}
}

func (s *analyticsService) GetActivityLogs(ctx context.Context, workspaceID string, channel, status *string, page, pageSize int32) (*domain.ActivityLogResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	g, gctx := errgroup.WithContext(ctx)

	var logs []domain.ActivityLog
	var totalCount int64

	g.Go(func() error {
		var err error
		offset := (page - 1) * pageSize
		logs, err = s.repo.ListActivityLogs(gctx, workspaceID, channel, status, pageSize, offset)
		return err
	})

	g.Go(func() error {
		var err error
		totalCount, err = s.repo.CountActivityLogs(gctx, workspaceID, channel, status)
		return err
	})

	if err := g.Wait(); err != nil {
		s.log.Error().Err(err).Msg("Failed to fetch activity logs")
		return nil, err
	}

	totalPages := int32((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return &domain.ActivityLogResponse{
		Logs:        logs,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}, nil
}


func findMostUsed(counts map[string]int64) string {
	var winner string
	var highestCount int64

	for k, val := range counts {
		if val > highestCount {
			highestCount = val
			winner = k
		}
	}
	return winner
}

func calculateTrend(current, previous int64) float32 {
	if previous == 0 {
		return 0
	}
	return float32(current-previous) / float32(previous) * 100
}

func buildTrends(current, previous *domain.AggregateMetrics) map[string]float32 {
	fields := []struct {
		name     string
		current  int64
		previous int64
	}{
		{"sent", current.TotalSent, previous.TotalSent},
		{"failed", current.TotalFailed, previous.TotalFailed},
		{"delivered", current.TotalDelivered, previous.TotalDelivered},
		{"bounced", current.TotalBounced, previous.TotalBounced},
	}

	trends := make(map[string]float32, len(fields))
	for _, f := range fields {
		trends[f.name] = calculateTrend(f.current, f.previous)
	}
	return trends
}
