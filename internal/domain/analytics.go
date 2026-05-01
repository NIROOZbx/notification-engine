package domain

import "time"

type AggregateMetrics struct {
	TotalSent      int64
	TotalDelivered int64
	TotalFailed    int64
	TotalBounced   int64
	ChannelCounts  map[string]int64
	ProviderCounts map[string]int64
}

type TimeSeriesData struct {
	Date           time.Time
	TotalSent      int64
	TotalDelivered int64
	TotalFailed    int64
}

type ProviderHealth struct {
	Provider   string
	AvgLatency int32
	LastSync   time.Time
}

type AnalyticsResponse struct {
	Aggregate struct {
		TotalSent           int64            `json:"total_sent"`
		TotalDelivered      int64            `json:"total_delivered"`
		TotalFailed         int64            `json:"total_failed"`
		TotalBounced        int64            `json:"total_bounced"`
		MostUsedChannel     string           `json:"most_used_channel"`
		MostUsedProvider    string           `json:"most_used_provider"`
		MostRecentProvider  string           `json:"most_recent_provider"`
		Trends             map[string]float32 `json:"trends"`
	} `json:"aggregate"`
	Channels   map[string]int64    `json:"channels"`
	Providers  []ProviderCount     `json:"providers"`
	TimeSeries []TimeSeriesDataDto `json:"time_series"`
	Health     struct {
		AverageLatencyMs int32            `json:"average_latency_ms"`
		LatencyTrend    []int32          `json:"latency_trend"`
		ActiveProviders []ProviderHealth `json:"active_providers"`
	} `json:"health"`
}

type ProviderCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type TimeSeriesDataDto struct {
	Label          string `json:"label"`
	SentCount      int64  `json:"sent_count"`
	DeliveredCount int64  `json:"delivered_count"`
	FailedCount    int64  `json:"failed_count"`
}

type ActivityLog struct {
	ID                string     `json:"id"`
	Channel           string     `json:"channel"`
	DeliveryStatus    string     `json:"delivery_status"`
	Recipient         string     `json:"recipient"`
	Provider          string     `json:"provider"`
	ProviderMessageID string     `json:"provider_message_id"`
	ProviderResponse  string     `json:"provider_response"`
	ErrorMessage      string     `json:"error_message"`
	TemplateID        string     `json:"template_id"`
	TemplateName      string     `json:"template_name"`
	ExternalUserID    string     `json:"external_user_id"`
	TriggerData       any        `json:"trigger_data"`
	DurationMs        int32      `json:"duration_ms"`
	AttemptCount      int32      `json:"attempt_count"`
	CreatedAt         time.Time  `json:"created_at"`
	SentAt            *time.Time `json:"sent_at"`
	DeliveredAt       *time.Time `json:"delivered_at"`
	FailedAt          *time.Time `json:"failed_at"`
}

type ActivityLogResponse struct {
	Logs        []ActivityLog `json:"logs"`
	TotalCount  int64         `json:"total_count"`
	TotalPages  int32         `json:"total_pages"`
	CurrentPage int32         `json:"current_page"`
	PageSize    int32         `json:"page_size"`
}

