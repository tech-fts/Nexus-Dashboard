package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/akz/nexus-dashboard/backend/internal/models"
	"github.com/akz/nexus-dashboard/backend/pkg/logger"
)

// Consumer wraps a Kafka consumer group reader.
type Consumer struct {
	reader *kafka.Reader
	log    *logger.Logger
}

// NewConsumer creates a new Kafka consumer group reader.
func NewConsumer(brokers []string, topic, groupID string, log *logger.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    10,                 // 10 bytes
		MaxBytes:    10e6,               // 10 MB
		MaxWait:     500 * time.Millisecond,
		CommitInterval: time.Second,
		StartOffset: kafka.LastOffset,
	})

	log.Info("kafka consumer initialized",
		"brokers", brokers,
		"topic", topic,
		"group_id", groupID,
	)

	return &Consumer{reader: reader, log: log}
}

// ReadMessage blocks until a message is available or context is cancelled.
func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// DecodeMetric parses a Kafka message value into a Metric model.
func DecodeMetric(msg kafka.Message) (*models.Metric, error) {
	var m models.Metric
	if err := json.Unmarshal(msg.Value, &m); err != nil {
		return nil, fmt.Errorf("decode metric: %w", err)
	}
	// Set ID from Kafka offset if not provided
	if m.ID == "" {
		m.ID = fmt.Sprintf("%s-%d", msg.Topic, msg.Offset)
	}
	return &m, nil
}

// Close cleanly shuts down the consumer.
func (c *Consumer) Close() error {
	c.log.Info("closing kafka consumer")
	return c.reader.Close()
}
