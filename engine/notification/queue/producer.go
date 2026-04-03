package queue

import (
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
)

type producer struct {
	writer *kafka.Writer
}

func NewProducer(brokerAddr string) *producer {
	return &producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddr),
			Balancer: &kafka.LeastBytes{},
			BatchSize: 100,
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *producer) Publish(ctx context.Context, topic string, event any) error {

	bytes,err:=sonic.Marshal(event)

	if err!=nil{
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Topic: topic,Value: bytes})
}

func (p *producer) Close() error {
	return p.writer.Close()
}