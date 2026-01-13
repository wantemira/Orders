package producer

import (
	"context"
	"encoding/json"
	"producer-service/internal/config"
	"producer-service/internal/kafka/messaging"
)

const (
	countMsg = 10
)

func createValidJSON() ([]byte, error) {
	order := createRandomOrder()
	return json.Marshal(order)
}

func ExternalSend(config config.KafkaConfig) {
	producer := messaging.NewKafkaProducer([]string{config.KafkaURL})
	defer func() {
		if err := producer.Close(); err != nil {
			config.Logger.Errorf("failed to close producer: %v", err)
		}
	}()
	for range countMsg {
		jsonBytes, err := createValidJSON()
		if err != nil {
			config.Logger.Errorf("producer.ExternalSend: Failed to create JSON: %v", err)
			return
		}

		var test map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &test); err != nil {
			config.Logger.Errorf("producer.ExternalSend: Generated JSON is invalid: %v", err)
			return
		}

		msg := messaging.Message{
			Key:   []byte("test-key"),
			Value: jsonBytes,
		}

		if err := producer.ProduceMessage(context.Background(), config.Topic, msg); err != nil {
			config.Logger.Errorf("producer.ExternalSend:  %v", err)
			return
		}

		config.Logger.Infof("producer.ExternalSend: Successfully sent message to topic %s", config.Topic)
	}
}
