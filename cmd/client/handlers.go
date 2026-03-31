package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		if outcome == gamelogic.MoveOutComeSafe {
			return pubsub.Ack
		} else if outcome == gamelogic.MoveOutcomeMakeWar {
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Println(err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(dw)
		switch outcome {
			case gamelogic.WarOutcomeNotInvolved:
				return pubsub.NackRequeue
			case gamelogic.WarOutcomeNoUnits:
				return pubsub.NackDiscard
			case gamelogic.WarOutcomeOpponentWon:
				err := pubsub.PublishGameLog(ch, dw.Attacker.Username, fmt.Sprintf("%s won a war against %s", winner, loser))
				if err != nil {
					return pubsub.NackRequeue
				}
				return pubsub.Ack
			case gamelogic.WarOutcomeYouWon:
				err := pubsub.PublishGameLog(ch, dw.Attacker.Username, fmt.Sprintf("%s won a war against %s", winner, loser))
				if err != nil {
					return pubsub.NackRequeue
				}
				return pubsub.Ack
			case gamelogic.WarOutcomeDraw:
				err := pubsub.PublishGameLog(ch, dw.Attacker.Username, fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser))
				if err != nil {
					return pubsub.NackRequeue
				}
				return pubsub.Ack
			default:
				fmt.Println("Unknown war outcome: ", outcome)
				return pubsub.NackDiscard
		}
	}
}
