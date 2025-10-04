package main

import (
	"fmt"

	"github.com/sidarun88/learn-pub-sub-starter/internal/gamelogic"
	"github.com/sidarun88/learn-pub-sub-starter/internal/pubsub"
	"github.com/sidarun88/learn-pub-sub-starter/internal/routing"
)

func handlerWarLogs() func(log routing.GameLog) pubsub.AckType {
	return func(log routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(log)
		if err != nil {
			fmt.Printf("error writing log: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
