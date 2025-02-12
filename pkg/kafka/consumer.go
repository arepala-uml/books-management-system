package kafka

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

// EventHandler processes events from Kafka
type EventHandler struct{}

// StartConsumer initializes a consumer and starts listening for events
func StartConsumer(brokerList []string) {
	// Create a new consumer group
	ctx := context.Background()
	consumer, err := sarama.NewConsumerGroup(brokerList, "book-events-group", nil)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}

	// Start consuming from the Kafka topic
	for {
		err := consumer.Consume(ctx, []string{"book_events"}, &EventHandler{})
		if err != nil {
			log.Printf("Error consuming message: %v", err)
		}
	}
}

// Setup is called when a consumer group is initialized
func (h *EventHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called when the consumer group is closed
func (h *EventHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from the Kafka topic
func (h *EventHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Consumed message: %s", string(message.Value))
		// You can implement further logic such as logging or analytics here
		sess.MarkMessage(message, "")
	}
	return nil
}
