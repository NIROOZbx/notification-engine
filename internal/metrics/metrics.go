package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Registry            *prometheus.Registry
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsTotal   *prometheus.CounterVec
	NotificationsSent   *prometheus.CounterVec
	NotificationsFailed *prometheus.CounterVec
	KafkaConsumerLag    *prometheus.GaugeVec
	ProviderDuration    *prometheus.HistogramVec
	KafkaProducerDuration *prometheus.HistogramVec
	KafkaMessageQueueTime *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()

	// 1. HTTP Request Duration
	httpRequestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "The duration of HTTP requests in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	// 2. Total HTTP Requests
	httpRequestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests processed.",
	}, []string{"method", "path", "status"})

	// 3. Notifications Sent
	notificationsSentTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "notifications_sent_total",
		Help: "Total number of notifications successfully sent.",
	}, []string{"channel", "provider"})

	// 4. Notifications Failed
	notificationsFailedTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "notifications_failed_total",
		Help: "Total number of notifications that failed to send.",
	}, []string{"channel", "provider"})

	// 5. Kafka Consumer Lag
	kafkaConsumerLag := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kafka_consumer_lag",
		Help: "Number of unprocessed messages in Kafka topic.",
	}, []string{"topic", "partition"})

	// 6. Provider Response Duration
	providerDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "provider_request_duration_seconds",
		Help:    "Time taken for external providers to respond.",
		Buckets: prometheus.DefBuckets,
	}, []string{"provider", "channel", "status"})

	kafkaProducerDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kafka_producer_duration_seconds",
		Help:    "Time taken to publish message to Kafka.",
		Buckets: prometheus.DefBuckets,
	}, []string{"topic", "status"})

	kafkaMessageQueueTime := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kafka_message_queue_time_seconds",
		Help:    "Time a message spent in the Kafka queue before being processed.",
		Buckets: prometheus.DefBuckets,
	}, []string{"topic", "channel"})

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		kafkaConsumerLag,
		httpRequestDuration,
		notificationsFailedTotal,
		notificationsSentTotal,
		httpRequestsTotal,
		providerDuration,
		kafkaProducerDuration,
		kafkaMessageQueueTime,
	)

	return &Metrics{
		Registry:            reg,
		HTTPRequestDuration: httpRequestDuration,
		HTTPRequestsTotal:   httpRequestsTotal,
		NotificationsSent:   notificationsSentTotal,
		NotificationsFailed: notificationsFailedTotal,
		KafkaConsumerLag:    kafkaConsumerLag,
		ProviderDuration:    providerDuration,
		KafkaProducerDuration: kafkaProducerDuration,
		KafkaMessageQueueTime: kafkaMessageQueueTime,
	}
}

func (m *Metrics) HTTPHandler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}
