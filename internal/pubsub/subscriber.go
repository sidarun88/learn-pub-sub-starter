package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	TransientQueue SimpleQueueType = iota
	DurableQueue
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %v", err)
	}

	args := amqp.Table{"x-dead-letter-exchange": "peril_dlx"}
	queue, err := ch.QueueDeclare(
		queueName,
		queueType != TransientQueue,
		queueType == TransientQueue,
		queueType == TransientQueue,
		false,
		args,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue: %v", err)
	}

	err = ch.QueueBind(
		queue.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %v", err)
	}

	return ch, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	messages, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = channel.Close()
		return err
	}

	go func() {
		defer channel.Close()
		for message := range messages {
			var data T
			err = json.Unmarshal(message.Body, &data)
			if err != nil {
				fmt.Printf("error unmarshalling JSON: %v", err)
				continue
			}

			switch handler(data) {
			case Ack:
				err = message.Ack(false)
				if err != nil {
					fmt.Printf("error acknowledging message: %v\n", err)
				}
			case NackRequeue:
				err = message.Nack(false, true)
				if err != nil {
					fmt.Printf("error nacking message: %v\n", err)
				}
			case NackDiscard:
				err = message.Nack(false, false)
				if err != nil {
					fmt.Printf("error nacking message: %v\n", err)
				}
			}
		}
	}()

	return nil
}
