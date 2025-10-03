package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, payload T) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
}
