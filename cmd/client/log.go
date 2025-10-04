package main

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sidarun88/learn-pub-sub-starter/internal/pubsub"
	"github.com/sidarun88/learn-pub-sub-starter/internal/routing"
)

func publishGameLog(pubChannel *amqp.Channel, username, message string) error {
	return pubsub.PublishGob(
		pubChannel,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.GameLogSlug, username),
		routing.GameLog{
			Username:    username,
			Message:     message,
			CurrentTime: time.Now(),
		},
	)
}
