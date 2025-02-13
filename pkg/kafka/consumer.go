package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/labstack/gommon/log"
)

type EventHandler struct{}

// Initializes a consumer and starts listening for events
func StartConsumer(brokerList []string, topic string) error {
	// Create a new consumer group
	log.Info(brokerList)
	consumer, err := sarama.NewConsumerGroup(brokerList, "book-events-group", nil)
	if err != nil {
		log.Errorf("Failed to start Kafka consumer: %v", err)
		return err
	}

	ctx := context.Background()
	for {
		err := consumer.Consume(ctx, []string{topic}, &EventHandler{})
		if err != nil {
			log.Infof("Error consuming message: %v", err)
		}
	}
}

func (h *EventHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Info("Consumer group setup called")
	return nil
}

func (h *EventHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Info("Consumer group cleanup called")
	return nil
}

func (h *EventHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Log the consumed message
		log.Infof("Consumed message: %s", string(message.Value))
		sess.MarkMessage(message, "")
	}
	return nil
}
