package main

import (
	"context"
	"log"
	"producer-service/internal/config"
	"producer-service/internal/kafka/messaging"
	"producer-service/internal/kafka/sender"
	"producer-service/pkg/closer"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	loggerLevelStr := config.GetEnv("LOGGER_LEVEL", logrus.DebugLevel.String())
	loggerLevel, err := logrus.ParseLevel(loggerLevelStr)
	if err != nil {
		logger.Errorf("main: %v", err)
	}
	logger.SetLevel(loggerLevel)
	logger.Infof("logrus: set %s level", loggerLevel)

	manager := closer.NewManager(logger)

	kafkaCfg, err := config.LoadKafkaConfig(logger)
	if err != nil {
		log.Printf("main: %v", err)
		return
	}
	logger.Info("main: kafka config Load - Success")
	producer := messaging.NewKafkaProducer([]string{kafkaCfg.KafkaURL})
	manager.Add(producer)

	orderSender := sender.NewOrderSender(producer, kafkaCfg.Topic, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go orderSender.Send(ctx)

	manager.WaitForSignal()

	cancel()

	time.Sleep(30 * time.Second)

	logger.Info("main: ExternalSend - Success")
}
