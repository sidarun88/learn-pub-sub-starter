package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Program running. Press Ctrl+C to exit gracefully.")

	sig := <-signalChan
	fmt.Printf("\nReceived signal: %v. Exiting...\n", sig)
}
