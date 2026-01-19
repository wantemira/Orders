package messaging

import (
	"context"
	"orders/pkg/models"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// isTemporaryError определяет временная ошибка или перманентная
func (c *KafkaConsumer) isTemporaryError(err error) bool {
	errStr := err.Error()

	// Временные ошибки (можно ретраить)
	temporaryPatterns := []string{
		"connection refused",
		"timeout",
		"deadline exceeded",
		"database is locked",
		"too many connections",
		"network error",
		"temporary failure",
		"try again later",
	}

	// Перманентные ошибки (не ретраить)
	permanentPatterns := []string{
		"validation failed",
		"invalid",
		"not found",
		"duplicate key",
		"already exists",
		"permission denied",
		"constraint",
	}

	for _, pattern := range temporaryPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	for _, pattern := range permanentPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return false
		}
	}

	return false
}

func (c *KafkaConsumer) handlePermanentErr(ctx context.Context, log *logrus.Entry, msg kafka.Message, errType string, err error) {
	log.WithFields(
		logrus.Fields{
			"error_type": errType,
			"error":      err.Error(),
			"message":    string(msg.Value),
		}).Error("Permanent error - messsage skipped")

	if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
		log.Errorf("Failed to commit invalid message: %v", commitErr)
	}
}
func (c *KafkaConsumer) handleProcessingErr(_ context.Context, log *logrus.Entry, msg kafka.Message, order *models.OrderJSON, err error, attemps int) {
	log.WithFields(
		logrus.Fields{
			"order_uid":    order.OrderUID,
			"error":        err.Error(),
			"attemps":      attemps,
			"message_size": len(msg.Value),
			"is_temporary": c.isTemporaryError(err),
		}).Error("Failed to proccess order after retries")
}
