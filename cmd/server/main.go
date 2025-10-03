package main

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sidarun88/learn-pub-sub-starter/internal/gamelogic"
	"github.com/sidarun88/learn-pub-sub-starter/internal/pubsub"
	"github.com/sidarun88/learn-pub-sub-starter/internal/routing"
)

const amqpURI = "amqp://guest:guest@localhost:5672/"

func main() {
	log.Println("Starting Peril server...")

	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	defer conn.Close()
	log.Println("Connected to RabbitMQ")

	pubChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	defer pubChannel.Close()
	log.Println("Opened a publisher channel")

	routingKey := routing.GameLogSlug + ".*"
	channel, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routingKey,
		pubsub.DurableQueue,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %s", err)
	}

	defer channel.Close()
	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		command := words[0]
		switch command {
		case "help":
			gamelogic.PrintServerHelp()
		case "quit":
			log.Println("Stopping server...")
			os.Exit(0)
		case "pause":
			log.Println("Sending pause message...")
			state := routing.PlayingState{IsPaused: true}
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				state,
			)
			if err != nil {
				log.Printf("Failed to publish 'pause' message: %v\n", err)
			}
			log.Println("Pause message sent")
		case "resume":
			log.Println("Sending resume message...")
			state := routing.PlayingState{IsPaused: false}
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.PauseKey,
				state,
			)
			if err != nil {
				log.Printf("Failed to publish 'resume' message: %v\n", err)
			}
			log.Println("Resume message sent")
		default:
			log.Printf("Unknown command: %s", command)
		}
	}
}
