package main

import (
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

const connUrl = "amqp://guest:guest@localhost:5672/"

func main() {
	conn, err := amqp.Dial(connUrl)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer conn.Close()

	log.Printf("Connention was successful")

	ch, err := conn.Channel()
	pubsub.PublishJSON(
		ch,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{IsPaused: true},
	)
	if err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Printf("Connection is closing")
}
