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
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	defer conn.Close()
	log.Println("Connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("failed to get username: %v", err)
	}

	gameState := gamelogic.NewGameState(username)
	queueName := routing.PauseKey + "." + gameState.GetUsername()
	err = pubsub.SubscribeJSON[routing.PlayingState](
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.TransientQueue,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("failed to subscribe to queue: %v", err)
	}

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
				fmt.Println(err)
			}
		case "move":
			_, err = gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
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
