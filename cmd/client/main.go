package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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

	pubChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open publish channel: %v", err)
	}
	defer pubChannel.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("failed to get username: %v", err)
	}

	gameState := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON[routing.PlayingState](
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, gameState.GetUsername()),
		routing.PauseKey,
		pubsub.TransientQueue,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("failed to subscribe to game state queue: %v", err)
	}

	err = pubsub.SubscribeJSON[gamelogic.ArmyMove](
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, gameState.GetUsername()),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.TransientQueue,
		handlerMove(pubChannel, gameState),
	)
	if err != nil {
		log.Fatalf("failed to subscribe to game moves queue: %v", err)
	}

	err = pubsub.SubscribeJSON[gamelogic.RecognitionOfWar](
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.DurableQueue,
		handlerWar(pubChannel, gameState),
	)
	if err != nil {
		log.Fatalf("failed to subscribe to game war queue: %v", err)
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
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				pubChannel,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, gameState.GetUsername()),
				move,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(move.Units), move.ToLocation)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(words) < 2 {
				fmt.Println("usage: spam <number>")
				continue
			}

			n, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Printf("error converting %s to int: %v, usage: spam <number>\n", words[1], err)
				continue
			}

			for _ = range n {
				msg := gamelogic.GetMaliciousLog()
				err = publishGameLog(pubChannel, username, msg)
				if err != nil {
					fmt.Printf("failed to publish log: %v\n", err)
				}
			}
		case "quit":
			gamelogic.PrintQuit()
			os.Exit(0)
		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}
