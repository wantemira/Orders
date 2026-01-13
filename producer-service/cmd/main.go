package main

import (
	"log"
	"producer-service/internal/config"
	"producer-service/internal/kafka/producer"

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
	logger.Infof("LOGRUS: SET %s Level", loggerLevel)

	kafkaCfg, err := config.LoadKafkaConfig(logger)
	if err != nil {
		log.Printf("main: %v", err)
		return
	}
	logger.Info("main: Kafka config Load - Success")

	producer.ExternalSend(*kafkaCfg)
	logger.Info("main: ExternalSend - Success")
}
