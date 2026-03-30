package main

import (
	"fmt"
	"os"
	"os/signal"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")

	connection := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connection)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	publishCh, err := conn.Channel()
	if err != nil {
		fmt.Println("Could not create channel")
	}

	fmt.Println("Connected to Peril server.")
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
		return
	}
	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"war",
		"war.#",
		pubsub.SimpleQueueDurable,
		handlerWar(gameState),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, publishCh),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
			case "spawn":
				err = gameState.CommandSpawn(words)
				if err != nil {
					fmt.Println(err)
					continue
				}
			case "move":
				mv, err := gameState.CommandMove(words)
				if err != nil {
					fmt.Println(err)
					continue
				}

				err = pubsub.PublishJSON(
					publishCh,
					routing.ExchangePerilTopic,
					routing.ArmyMovesPrefix+"."+mv.Player.Username,
					mv,
				)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)
			case "status":
				gameState.CommandStatus()
			case "help":
				gamelogic.PrintClientHelp()
			case "spam":
				fmt.Println("Spamming not allowed yet!")
			case "quit":
				gamelogic.PrintQuit()
				return
			default:
				fmt.Println("Unknown command")
		}
	}

	// Wait for Ctrl+C
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nExiting...")
}
