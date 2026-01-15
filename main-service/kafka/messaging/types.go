// Package messaging содержит клиенты для работы с Kafka
package messaging

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Message — сообщение
type Message struct {
	Key   []byte
	Value []byte
}

// Producer определяет интерфейс для отправки сообщений в Kafka
type Producer interface {
	ProduceMessage(ctx context.Context, topic string, msg Message) error
	Close() error
}

// Handler обрабатывает сообщения из Kafka
type Handler func(ctx context.Context, msg Message) error

// Consumer определяет интерфейс для чтения сообщений из Kafka
type Consumer interface {
	Run(ctx context.Context)
	ConsumeMessage(ctx context.Context) error
	Commit(ctx context.Context, msg kafka.Message) error
	Close(ctx context.Context) error
	Name() string
}
