package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

var producer sarama.SyncProducer

// InitProducer initializes the Kafka producer
func InitProducer(brokerList []string) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	var err error
	producer, err = sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatalf("Failed to start Kafka producer: %v", err)
		return err
	}

	log.Println("Kafka producer initialized")
	return nil
}

// PublishEvent publishes an event to the specified Kafka topic
func PublishEvent(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	_, _, err := producer.SendMessage(msg)
	if err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		return err
	}

	log.Printf("Event sent to Kafka topic '%s'", topic)
	return nil
}

// CloseProducer gracefully shuts down the Kafka producer
func CloseProducer() {
	if err := producer.Close(); err != nil {
		log.Fatalf("Failed to close Kafka producer: %v", err)
	}
	log.Println("Kafka producer closed")
}
