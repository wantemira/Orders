// Package sender содержит логику отправки заказов в Kafka
package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"producer-service/internal/kafka/messaging"
	"time"

	"github.com/sirupsen/logrus"
)

// OrderSender отправляет сгенерированные заказы в Kafka
type OrderSender struct {
	producer messaging.Producer
	logger   *logrus.Logger
	topic    string
}

// NewOrderSender создает новый экземпляр OrderSender
func NewOrderSender(producer messaging.Producer, topic string, logger *logrus.Logger) *OrderSender {
	return &OrderSender{
		producer: producer,
		logger:   logger,
		topic:    topic,
	}
}

// Send запускает периодическую отправку заказов
func (s *OrderSender) Send(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Error("sender.Send: stop to send messages")
			return
		case <-ticker.C:
			if err := s.SendOnes(ctx); err != nil {
				s.logger.Errorf("sender.Send: failed to send: %v", err)
			}
		}
	}
}

// SendOnes отправляет один сгенерированный заказ
func (s *OrderSender) SendOnes(ctx context.Context) error {
	order := createRandomOrder()
	jsonBytes, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("sender.Send: Failed to create JSON: %v", err)

	}

	var test map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &test); err != nil {
		return fmt.Errorf("sender.Send: Generated JSON is invalid: %v", err)
	}

	msg := messaging.Message{
		Key:   []byte(order.OrderUID),
		Value: jsonBytes,
	}

	if err := s.producer.ProduceMessage(context.Background(), s.topic, msg); err != nil {
		return fmt.Errorf("sender.Send:  %v", err)
	}

	s.logger.Infof("sender.Send: Successfully sent message to topic %s", s.topic)
	return nil
}
