// Package messaging содержит клиенты для работы с Kafka
package messaging

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// kafkaProducer реализует Producer для отправки сообщений в Kafka
type kafkaProducer struct {
	writer *kafka.Writer
	name   string
}

// NewKafkaProducer создает новый экземпляр kafkaProducer
func NewKafkaProducer(brokers []string) Producer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		BatchTimeout:           1 * time.Second,
		WriteTimeout:           10 * time.Second,
	}
	return &kafkaProducer{
		writer: writer,
		name:   "kafka producer",
	}
}

// ProduceMessage отправляет сообщение в указанный топик Kafka
func (p *kafkaProducer) ProduceMessage(ctx context.Context, topic string, msg Message) error {
	kafkaMsg := kafka.Message{
		Topic: topic,
		Key:   msg.Key,
		Value: msg.Value,
	}
	return p.writer.WriteMessages(ctx, kafkaMsg)
}

// Close закрывает соединение с Kafka
func (p *kafkaProducer) Close(ctx context.Context) error {
	return p.writer.Close()
}

// Name возвращает имя ресурса для логирования
func (p *kafkaProducer) Name() string { return p.name }
