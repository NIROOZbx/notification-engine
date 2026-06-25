package queue

import (
	"context"
	"strconv"
	"time"

	"github.com/NIROOZbx/notification-engine/engine/notification/models"
	"github.com/NIROOZbx/notification-engine/internal/metrics"
	"github.com/NIROOZbx/notification-engine/pkg/serializer"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

type ProcessFunc func(ctx context.Context, event *models.NotificationEvent) error

type Consumer interface {
	Start(ctx context.Context) error
	Close() error
}

type consumer struct {
	reader      *kafka.Reader
	processFunc ProcessFunc
	log         zerolog.Logger
	metrics     *metrics.Metrics
}

func NewConsumer(brokerAddr string, topic string, groupID string, fn ProcessFunc, log zerolog.Logger, metrics *metrics.Metrics) *consumer {
	return &consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Topic:   topic,
			Brokers: []string{brokerAddr},
			GroupID: groupID,
			

		}),
		processFunc: fn,
		log:         log,
		metrics:     metrics,
	}
}

func (c *consumer) Start(ctx context.Context) error {

	for {
		msg, err := c.reader.FetchMessage(ctx)

		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.log.Error().Err(err).Msg("failed to fetch message")
			continue
		}
		event := &models.NotificationEvent{}
		if err := serializer.Unmarshal(msg.Value, event); err != nil {
			c.log.Error().Err(err).Msg("failed to unmarshal notification event")
			c.reader.CommitMessages(ctx, msg)
			continue
		}

		c.log.Info().
			Str("event_id", event.NotificationLogID).
			Int("attempt", event.AttemptNumber+1).
			Msg("consumer picked up message")

		if err:=c.processFunc(ctx, event); err!=nil{
			c.log.Error().Err(err).Str("event_id", event.NotificationLogID).Msg("failed to process event")
			continue 
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Error().Err(err).Msg("failed to commit message")
		}

		if c.metrics != nil {
			stats := c.reader.Stats()
			c.metrics.KafkaConsumerLag.WithLabelValues(msg.Topic, strconv.Itoa(msg.Partition)).Set(float64(stats.Lag))

			if event.PublishedAt > 0 {
				queueTime := float64(time.Now().UnixNano()-event.PublishedAt) / float64(time.Second)
				c.metrics.KafkaMessageQueueTime.WithLabelValues(msg.Topic, event.Channel).Observe(queueTime)
			}
		}
	}
}

func (c *consumer) Close() error {
	return c.reader.Close()
}
