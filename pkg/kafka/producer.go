package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

// Every POST, PUT, DELETE request should publish an event to a Kafka topic
func PublishEvent(topic string, message []byte) error {
	brokersUrl := []string{fmt.Sprintf("%s:%s", viper.GetString("KAFKA_HOST"), viper.GetString("KAFKA_PORT"))}

	// Create the Kafka producer
	producer, err := ConnectProducer(brokersUrl)
	if err != nil {
		return err
	}
	defer producer.Close()
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	// Send the message to Kafka
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Errorf("Error sending message: %v", err)
		return err
	}
	fmt.Printf("Message sent to topic %s, partition %d, offset %d\n", topic, partition, offset)
	return nil
}

// Creates and returns a Kafka producer
func ConnectProducer(brokersUrl []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	// Create a new producer instance
	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
