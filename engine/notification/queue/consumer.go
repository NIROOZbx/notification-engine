package queue

import (
	"context"

	"github.com/NIROOZbx/notification-engine/engine/notification/models"
	"github.com/bytedance/sonic"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

type ProcessFunc func(ctx context.Context, event models.NotificationEvent) error

type consumer struct {
	reader      *kafka.Reader
	processFunc ProcessFunc
	log         zerolog.Logger
}

func NewConsumer(brokerAddr string, topic string, fn ProcessFunc, log zerolog.Logger) *consumer {
	return &consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Topic:   topic,
			Brokers: []string{brokerAddr},
			GroupID: "notification-engine",
			

		}),
		processFunc: fn,
		log:         log,
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
		var event models.NotificationEvent

		if err := sonic.Unmarshal(msg.Value, &event); err != nil {
			c.log.Error().Err(err).Msg("failed to unmarshal notification event")
			c.reader.CommitMessages(ctx, msg)
			continue
		}

		if err:=c.processFunc(ctx,event); err!=nil{
			c.log.Error().Err(err).Str("event_id", event.NotificationLogID).Msg("failed to process event")
			continue 
		}

		if err:=c.reader.CommitMessages(ctx,msg); err!=nil{
			 c.log.Error().Err(err).Msg("failed to commit message")
		}
	}

}

func (c *consumer) Close() error {
	return c.reader.Close()
}
