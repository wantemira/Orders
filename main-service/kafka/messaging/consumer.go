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
	reader  *kafka.Reader
	logger  *logrus.Logger
	handler *subs.Handler
}

// NewKafkaConsumer создает новый экземпляр KafkaConsumer
func NewKafkaConsumer(brokers []string, topic string, groupID string, logger *logrus.Logger, handler *subs.Handler) Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})

	return &KafkaConsumer{
		reader:  reader,
		logger:  logger,
		handler: handler,
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
		log.Errorf("KafkaConsumer.ConsumeMessage: Failed to unmarshal message: %v. Message %s", err, string(kafkaMsg.Value))

		if commitErr := c.reader.CommitMessages(ctx, kafkaMsg); commitErr != nil {
			log.Errorf("KafkaConsumer.ConsumeMessage: Failed to commit invalid message: %v", commitErr)
			return fmt.Errorf("commit invalid message: %w", commitErr)
		}
		log.Warn("KafkaConsumer.ConsumeMessage: Invalid message skipped and committed")
		return nil
	}
	validate := validator.New()
	if err := validate.Struct(order); err != nil {
		log.Errorf("KafkaConsumer.ConsumeMessage: Validation failed: %v", err)
		if err := c.reader.CommitMessages(ctx, kafkaMsg); err != nil {
			log.Errorf("KafkaConsumer.ConsumeMessage: failed to commit invalid message: %v", err)
			return fmt.Errorf("commit invalid message: %w", err)
		}
		log.Warn("KafkaConsumer.ConsumeMessage: Invalid data skipped and commited")
		return nil
	}

	log = log.WithField("order_uid", order.OrderUID)
	log.Info("KafkaConsumer.Run: Read Message")
	if err = c.handler.Create(ctx, order); err != nil {
		log.Errorf("KafkaConsumer.ConsumeMessage: %v", err)
		return fmt.Errorf("failed to create order %s. error: %w", order.OrderUID, err)

	}
	if err = c.Commit(ctx, kafkaMsg); err != nil {
		log.Errorf("KafkaConsumer.ConsumeMessage: %v", err)
		return fmt.Errorf("failed to commit message %s. error: %w", order.OrderUID, err)

	}
	log.Info("KafkaConsumer.Run: Commit Message")
	return nil
}

// Commit подтверждает обработку сообщения в Kafka
func (c *KafkaConsumer) Commit(ctx context.Context, msg kafka.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

// Close закрывает соединение с Kafka
func (c *KafkaConsumer) Close() error {
	c.logger.Info("KafkaConsumer.Close: Closing Kafka consumer")
	return c.reader.Close()
}
