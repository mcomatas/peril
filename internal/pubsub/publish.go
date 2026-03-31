package pubsub

import (
	"encoding/json"
	"encoding/gob"
	"context"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"time"
	"bytes"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	marshalVal, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        marshalVal,
	})
	if err != nil {
		return err
	}

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	if err := enc.Encode(val); err != nil {
		return err
	}

	err := ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        network.Bytes(),
	})
	if err != nil {
		return err
	}

	return nil
}

func PublishGameLog(ch *amqp.Channel, username, message string) error {
	return PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			CurrentTime: time.Now(),
			Message:	 message,
			Username: 	 username,
		},
	)
}
