// Package main является точкой входа в приложение
package main

import (
	"context"
	"fmt"
	"orders/internal/config"
	"orders/internal/database"
	"orders/internal/subs"
	"orders/kafka/messaging"
	"orders/router"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type Application struct {
	logger      *logrus.Logger
	db          *pgx.Conn // unused / check mb delete field
	DBHandler   *database.HandlerDB
	subsService *subs.Service
	subsHandler *subs.Handler
	server      *router.Server
	consumer    messaging.Consumer
	PostgresCfg *config.PostgresConfig
	kafkaCfg    *config.KafkaConfig
}

func main() {
	app, err := setupApplication()
	if err != nil {
		app.logger.Fatalf("main: Error with Setup Application")
	}
	// CACHE
	if err := app.subsService.WarmUpCache(context.Background()); err != nil {
		app.logger.Warnf("main: [CACHE]: Failed to warm up: %v", err)
	}
	app.logger.Infof("main: [CACHE]: Warm up")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	defer stop()

	// SERVER
	go app.server.Run()

	// KAFKA CONSUMER
	defer func() {
		if err := app.consumer.Close(); err != nil {
			app.logger.Errorf("failed to close consumer: %v", err)
		}
	}()

	app.logger.Infof("main: [KAFKA_CONSUMER]: Run")
	app.consumer.Run(ctx)

	<-ctx.Done()
	app.logger.Info("[GLOBAL]: Service stopped..")
}

func setupApplication() (*Application, error) {
	logger := setupLogger()

	postgresCfg, err := config.LoadPostgresConfig(logger) // "postgres://postgres:password@postgres:5432/Orders"
	if err != nil {
		logger.Errorf("main.setupApplication: [POSTGRES]: Error with load config: %v", err)
		// return nil, fmt.Errorf("main.setupApplication: %v", err)
	}
	URL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", postgresCfg.User, postgresCfg.Password, postgresCfg.Host, postgresCfg.Port, postgresCfg.Name)
	logger.Infof("main.setupApplication: [POSTGRES] Config was Load: %+v\n URL: %s", postgresCfg, URL)
	conn, dbHandler, err := setupDatabase(URL, logger)
	if err != nil {
		logger.Errorf("main.setupApplication: [POSTGRES]: Error with setup db: %v", err)
		return nil, fmt.Errorf("main.setupApplication: %v", err)
	}
	cache := subs.NewInMemoryCache(logger)
	subsRepo := subs.NewRepository(conn, logger)
	subsService := subs.NewService(subsRepo, logger, cache)
	subsHandler := subs.NewHandler(subsService, logger)

	kafkaCfg, err := config.LoadKafkaConfig(logger)
	if err != nil {
		logger.Errorf("main.setupApplication:: [KAFKA]: Error with load config: %v", err)
		// return nil, fmt.Errorf("main.setupApplication: %v", err)
	}
	logger.Infof("main.setupApplication:: [KAFKA]: Config was Load: %+v", kafkaCfg)
	kafkaConsumer := messaging.NewKafkaConsumer([]string{kafkaCfg.KafkaURL}, kafkaCfg.Topic, kafkaCfg.GroupConsumer, logger, subsHandler)

	server := router.NewServer(subsHandler, logger)
	return &Application{
		logger:      logger,
		db:          conn,
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

	loggerLevelStr := config.GetEnv("LOGGER_LEVEL", logrus.DebugLevel.String())

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
