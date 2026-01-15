// Package messaging содержит клиенты для работы с Kafka
package messaging

import (
	"context"
)

// Message представляет сообщение Kafka
type Message struct {
	Key   []byte
	Value []byte
}

// Producer определяет интерфейс для отправки сообщений в Kafka
type Producer interface {
	ProduceMessage(ctx context.Context, topic string, msg Message) error
	Close(ctx context.Context) error
	Name() string
}
