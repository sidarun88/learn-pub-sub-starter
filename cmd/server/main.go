package main

import (
	"fmt"
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
		log.Fatalf("failed to connect to RabbitMQ: %s", err)
	}

	defer conn.Close()
	log.Println("Connected to RabbitMQ")

	pubChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open publish channel: %s", err)
	}

	defer pubChannel.Close()
	log.Println("Opened the publisher channel")

	err = pubsub.SubscribeGob[routing.GameLog](
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		fmt.Sprintf("%s.*", routing.GameLogSlug),
		pubsub.DurableQueue,
		handlerWarLogs(),
	)
	if err != nil {
		log.Fatalf("failed to subscribe to game logs: %s", err)
	}

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
				pubChannel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				state,
			)
			if err != nil {
				log.Printf("failed to publish 'pause' message: %v\n", err)
				continue
			}
			log.Println("Pause message sent")
		case "resume":
			log.Println("Sending resume message...")
			state := routing.PlayingState{IsPaused: false}
			err = pubsub.PublishJSON(
				pubChannel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				state,
			)
			if err != nil {
				log.Printf("failed to publish 'resume' message: %v\n", err)
				continue
			}
			log.Println("Resume message sent")
		default:
			log.Printf("Unknown command: %s", command)
		}
	}
}
