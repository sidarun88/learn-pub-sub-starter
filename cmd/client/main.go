package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	ch, queue, err := pubsub.DeclareAndBind(
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
	log.Printf("Declared and bounded queue '%s'\n", queue.Name)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Program running. Press Ctrl+C to exit gracefully.")

	<-signalChan
	log.Println("Exiting...")
}
