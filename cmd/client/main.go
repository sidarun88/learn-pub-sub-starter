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
	log.Println("Starting Peril client...")

	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	defer conn.Close()
	log.Println("Connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Failed to get username: %s", err)
	}

	queueName := routing.PauseKey + "." + username
	ch, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.TransientQueue,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %s", err)
	}

	defer ch.Close()
	gameState := gamelogic.NewGameState(username)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		command := words[0]
		switch command {
		case "spawn":
			err = gameState.CommandSpawn(words)
			if err != nil {
				fmt.Printf("Failed to spawn: %s\n", err)
			}
		case "move":
			_, err = gameState.CommandMove(words)
			if err != nil {
				fmt.Printf("Failed to move: %s\n", err)
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			os.Exit(0)
		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}
