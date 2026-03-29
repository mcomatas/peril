package pubsub

import (
	"fmt"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T),
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
			handler(data)
			msg.Ack(false)
		}
	}()

	return nil
}
