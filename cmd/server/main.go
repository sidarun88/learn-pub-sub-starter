package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sidarun88/learn-pub-sub-starter/internal/pubsub"
	"github.com/sidarun88/learn-pub-sub-starter/internal/routing"
)

const amqpURI = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")

	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	defer conn.Close()
	fmt.Println("Connected to RabbitMQ")

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	state := routing.PlayingState{IsPaused: true}
	err = pubsub.PublishJSON(
		channel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		state,
	)
	if err != nil {
		log.Printf("Failed to publish a message: %v\n", err)
	}
	fmt.Println("Published 'Pause' message")
}
