// Package main является точкой входа в приложение
package main

import (
	"context"
	"fmt"
	"orders/internal/config"
	"orders/internal/database"
	"orders/internal/subs"
	"orders/kafka/messaging"
	"orders/pkg/closer"
	utilsCfg "orders/pkg/config"
	"orders/router"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type Application struct {
	DBHandler   *database.HandlerDB
	subsService *subs.Service
	subsHandler *subs.Handler
	server      *router.Server
	consumer    messaging.Consumer
	PostgresCfg *config.PostgresConfig
	kafkaCfg    *config.KafkaConfig
}

func main() {
	logger := setupLogger()
	manager := closer.NewManager(logger)

	app, err := setupApplication(logger, manager)
	if err != nil {
		logger.Fatalf("main: Error with Setup Application")
	}
	// Cache
	if err := app.subsService.WarmUpCache(context.Background()); err != nil {
		logger.Warnf("main: [CACHE]: Failed to warm up: %v", err)
	}
	logger.Infof("main: [CACHE]: Warm up")

	// Server
	logger.Infof("main: [HTTP SERVER]: Run")
	go app.server.Run()

	// Kafka
	logger.Infof("main: [KAFKA CONSUMER]: Run")
	go app.consumer.Run(context.Background())

	manager.WaitForSignal()
	logger.Info("[GLOBAL]: Service stopped..")
}

func setupApplication(logger *logrus.Logger, manager *closer.Manager) (*Application, error) {

	postgresCfg, err := config.LoadPostgresConfig(logger) // "postgres://postgres:password@postgres:5432/Orders"
	if err != nil {
		return nil, fmt.Errorf("load postgres config: %w", err)
	}

	URL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", postgresCfg.User, postgresCfg.Password, postgresCfg.Host, postgresCfg.Port, postgresCfg.Name)
	logger.Infof("main.setupApplication: [POSTGRES] Config was Load: %+v\n URL: %s", postgresCfg, URL)
	conn, dbHandler, err := setupDatabase(URL, logger)
	if err != nil {
		return nil, fmt.Errorf("setup database: %w", err)
	}
	manager.Add(dbHandler)

	cache := subs.NewInMemoryCache(logger)
	subsRepo := subs.NewRepository(conn, logger)
	subsService := subs.NewService(subsRepo, logger, cache)
	subsHandler := subs.NewHandler(subsService, logger)

	kafkaCfg, err := config.LoadKafkaConfig(logger)
	if err != nil {
		return nil, fmt.Errorf("load kafka config: %w", err)
	}

	maxRetries := 3
	kafkaConsumer := messaging.NewKafkaConsumer([]string{kafkaCfg.KafkaURL}, kafkaCfg.Topic, kafkaCfg.GroupConsumer, logger, subsHandler, maxRetries)
	manager.Add(kafkaConsumer)

	server := router.NewServer(subsHandler, logger)
	manager.Add(server)

	return &Application{
		DBHandler:   dbHandler,
		subsService: subsService,
		subsHandler: subsHandler,
		server:      server,
		consumer:    kafkaConsumer,
		PostgresCfg: postgresCfg,
		kafkaCfg:    kafkaCfg,
	}, nil

}

func setupLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	loggerLevelStr := utilsCfg.GetEnv("LOGGER_LEVEL", logrus.DebugLevel.String())

	loggerLevel, err := logrus.ParseLevel(loggerLevelStr)
	if err != nil {
		logger.SetLevel(logrus.DebugLevel)
		logger.Warnf("main.setupLogger: Failed to parse log level, using Info: %v", err)
	} else {
		logger.SetLevel(loggerLevel)
	}
	logger.Infof("main.setupLogger: Logger initialized with level: %s", logger.Level)
	return logger
}

func setupDatabase(URL string, logger *logrus.Logger) (*pgx.Conn, *database.HandlerDB, error) {

	conn, err := pgx.Connect(context.Background(), URL)
	if err != nil {
		return nil, nil, fmt.Errorf("main.setupDatabase: failed to connect to database: %w", err)
	}
	if err := conn.Ping(context.Background()); err != nil {
		return nil, nil, fmt.Errorf("main.setupDatabase: database ping failed: %w", err)
	}
	logger.Infof("main: [PGX]: Connected")

	handlerDB := database.NewHandlerDB(conn, logger)
	if err := handlerDB.CreateTables(context.Background(), conn); err != nil {
		return nil, nil, fmt.Errorf("main.setupDatabase: failed to create tables: %w", err)
	}
	logger.Info("main.setupDatabase: Database connection established and tables created")
	return conn, handlerDB, nil
}
