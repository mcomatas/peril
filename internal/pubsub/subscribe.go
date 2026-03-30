package pubsub

import (
	"fmt"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	msgs, err := channel.Consume(queue.Name, "", false, false, false, false, nil)

	go func() {
		for msg := range msgs {
			var data T
			err := json.Unmarshal(msg.Body, &data)
			if err != nil {
				fmt.Println(err)
				continue
			}
			ackType := handler(data)
			switch ackType {
				case Ack:
					fmt.Println("Ack")
					msg.Ack(false)
				case NackRequeue:
					fmt.Println("NackRequeue")
					msg.Nack(false, true)
				case NackDiscard:
					fmt.Println("NackDiscard")
					msg.Nack(false, false)
			}
		}
	}()

	return nil
}
