package queue

import (
	"context"
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