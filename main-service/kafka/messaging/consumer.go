// Package messaging содержит клиенты для работы с Kafka
package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"orders/internal/subs"
	"orders/pkg/models"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// KafkaConsumer реализует Consumer для чтения сообщений из Kafka
type KafkaConsumer struct {
	reader     *kafka.Reader
	logger     *logrus.Logger
	handler    *subs.Handler
	name       string
	maxRetries int
}

// NewKafkaConsumer создает новый экземпляр KafkaConsumer
func NewKafkaConsumer(brokers []string, topic string, groupID string, logger *logrus.Logger, handler *subs.Handler, maxRetries int) Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})

	return &KafkaConsumer{
		reader:     reader,
		logger:     logger,
		handler:    handler,
		name:       "kafka consumer",
		maxRetries: maxRetries,
	}
}

// Run запускает потребителя Kafka
func (c *KafkaConsumer) Run(ctx context.Context) {
	c.logger.Info("KafkaConsumer.Run: Starting consumer...")
	c.logger.Infof("KafkaConsumer.Run: Brokers: %v, Topic: %s, GroupID: %s",
		c.reader.Config().Brokers,
		c.reader.Config().Topic,
		c.reader.Config().GroupID)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("KafkaConsume.Run: Consumer stop (context canceled)")
			return
		default:
			if err := c.ConsumeMessage(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				c.logger.Errorf("KafkaConsumer: Error consuming message: %v", err)
				time.Sleep(2 * time.Second)
			}
		}
	}

}

// ConsumeMessage читает и обрабатывает сообщения из Kafka
func (c *KafkaConsumer) ConsumeMessage(ctx context.Context) error {
	kafkaMsg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		c.logger.Errorf("KafkaConsumer.ConsumeMessage: failed to fetch msg: %v", err)
		return fmt.Errorf("fetch message: %w", err)
	}
	log := c.logger.WithFields(logrus.Fields{
		"topic":     kafkaMsg.Topic,
		"partition": kafkaMsg.Partition,
		"offset":    kafkaMsg.Offset,
		"key":       kafkaMsg.Key,
	})
	log.Info("KafkaConsumer.Run: Received kafka message")
	var order *models.OrderJSON
	if err := json.Unmarshal(kafkaMsg.Value, &order); err != nil {
		c.handlePermanentErr(ctx, log, kafkaMsg, "json_unmarshal", err)
		return nil
	}
	validate := validator.New()
	if err := validate.Struct(order); err != nil {
		c.handlePermanentErr(ctx, log, kafkaMsg, "validation", err)
		return nil
	}

	log = log.WithField("order_uid", order.OrderUID)
	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		log.WithField("attempt", attempt).Info("Processing order")

		if err = c.handler.Create(ctx, order); err == nil {
			if commitErr := c.Commit(ctx, kafkaMsg); commitErr != nil {
				log.Errorf("Failed to commit after success: %v", commitErr)
			}
			log.Info("Order processed successfully")
			return nil
		}

		lastErr = err

		if c.isTemporaryError(err) && attempt < c.maxRetries {
			backoff := time.Duration(attempt) * time.Second
			log.WithField("backoff_seconds", attempt).Warnf("Temporary error, retrying %v", err)
			time.Sleep(backoff)
			continue
		}

		break
	}
	c.handleProcessingErr(ctx, log, kafkaMsg, order, lastErr, c.maxRetries)

	if commitErr := c.Commit(ctx, kafkaMsg); commitErr != nil {
		log.Errorf("Failed to commit after error %v", commitErr)
	}

	return nil
}

// Commit подтверждает обработку сообщения в Kafka
func (c *KafkaConsumer) Commit(ctx context.Context, msg kafka.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

// Close закрывает соединение с Kafka
func (c *KafkaConsumer) Close(ctx context.Context) error {
	c.logger.Info("KafkaConsumer.Close: Closing Kafka consumer")
	return c.reader.Close()
}

func (c *KafkaConsumer) Name() string { return c.name }
