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

	name, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		"pause."+name,
		routing.PauseKey,
		1,
	)
	if err != nil {
		log.Fatal(err)
	}

	gameState := gamelogic.NewGameState(name)
	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		"pause."+name,
		routing.PauseKey,
		1,
		handlerPause(gameState),
	)
loop:
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn":
			gameState.CommandSpawn(input)
		case "move":
			gameState.CommandMove(input)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Print("Spamming not allowed yet!\n")
		case "quit":
			gamelogic.PrintQuit()
			break loop
		default:
			fmt.Print("Invalid command\n")

		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Printf("Connection is closing")
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
