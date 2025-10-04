package pubsub

import (
	"bytes"
	"encoding/gob"
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

func declareAndBind(
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

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	unMarshaller func([]byte) (T, error),
) error {
	channel, queue, err := declareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	err = channel.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("could not set QoS: %v", err)
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
			data, err := unMarshaller(message.Body)
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	unMarshaller := func(jsonBytes []byte) (T, error) {
		var result T
		err := json.Unmarshal(jsonBytes, &result)
		return result, err
	}
	return subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		unMarshaller,
	)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	unMarshaller := func(gobBytes []byte) (T, error) {
		var result T
		buffer := bytes.NewBuffer(gobBytes)
		decoder := gob.NewDecoder(buffer)
		err := decoder.Decode(&result)
		return result, err
	}
	return subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		unMarshaller,
	)
}
