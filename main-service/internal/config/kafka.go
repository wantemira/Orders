// Package config предоставляет функции для загрузки конфигурации
package config

import (
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// KafkaConfig содержит конфигурацию для подключения к Kafka
type KafkaConfig struct {
	KafkaURL      string
	Topic         string
	GroupConsumer string
	Logger        *logrus.Logger
}

// LoadKafkaConfig загружает конфигурацию Kafka из переменных окружения
func LoadKafkaConfig(logger *logrus.Logger) (*KafkaConfig, error) {
	envPath := filepath.Join("configs", ".env")
	if err := godotenv.Load(envPath); err != nil {
		logger.Errorf("config.LoadPostgresConfig: %v", err)
		return nil, fmt.Errorf("config.LoadPostgresConfig: %w", err)
	}
	config := &KafkaConfig{
		KafkaURL:      GetEnv("KAFKA_URL", "kafka:9092"),
		Topic:         GetEnv("TEST_TOPIC", "test_topic"),
		GroupConsumer: GetEnv("GROUP_ID", "test_group"),
	}
	return config, nil
}
