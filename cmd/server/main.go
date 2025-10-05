package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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
	if err != nil {
		log.Fatal(err)
	}

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "pause":
			fmt.Print("Sending PAUSE msg")
			pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)

		case "resume":
			fmt.Print("Sending RESUME msg")
			pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
		case "quit":
			fmt.Print("Exiting")

		default:
			fmt.Print("Unknown command")
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Printf("Connection is closing")
}
