package config

import (
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// PostgresConfig содержит конфигурацию для подключения к PostgreSQL
type PostgresConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// LoadPostgresConfig загружает конфигурацию PostgreSQL из переменных окружения
func LoadPostgresConfig(logger *logrus.Logger) (*PostgresConfig, error) {
	envPath := filepath.Join("configs", ".env")
	if err := godotenv.Load(envPath); err != nil {
		logger.Errorf("config.LoadPostgresConfig: %v", err)
		return nil, fmt.Errorf("config.LoadPostgresConfig: %w", err)
	}
	config := &PostgresConfig{
		Host:     GetEnv("DB_HOST", "localhost"),
		Port:     GetEnv("DB_PORT", "5432"),
		Name:     GetEnv("DB_NAME", "Orders"),
		User:     GetEnv("DB_USER", "postgres"),
		Password: GetEnv("DB_PASSWORD", "password"),
	}
	return config, nil
}
