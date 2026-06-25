package queue

import (
	"context"
	"time"

	"github.com/NIROOZbx/notification-engine/internal/metrics"
	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
)

type producer struct {
	writer  *kafka.Writer
	metrics *metrics.Metrics
}

func NewProducer(brokerAddr string, metrics *metrics.Metrics) *producer {
	return &producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddr),
			Balancer: &kafka.LeastBytes{},
			BatchSize: 100,
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: kafka.RequireAll,
		},
		metrics: metrics,
	}
}

func (p *producer) Publish(ctx context.Context, topic string, event any) error {

	bytes,err:=sonic.Marshal(event)

	if err!=nil{
		return err
	}
	startTime := time.Now()
	err = p.writer.WriteMessages(ctx, kafka.Message{Topic: topic, Value: bytes})

	if p.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		p.metrics.KafkaProducerDuration.WithLabelValues(topic, status).Observe(time.Since(startTime).Seconds())
	}

	return err
}

func (p *producer) Close() error {
	return p.writer.Close()
}