package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	ctx := context.Background()
	err = ch.PublishWithContext(
		ctx,
		exchange,
		key,
		false,
		false,
		amqp.Publishing{ContentType: "application/json", Body: jsonData},
	)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

type SimpleQueueType int

const (
	durable SimpleQueueType = iota
	transient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable := true
	autoDelete := true
	exclusive := true
	noWait := false
	switch queueType {
	case 0:
		autoDelete = false
		exclusive = false
	case 1:
		durable = false
	}
	que, err := ch.QueueDeclare(queueName, durable, autoDelete, exclusive, noWait, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = ch.QueueBind(que.Name, key, exchange, noWait, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, que, nil
}

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
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	dlvCh, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for ele := range dlvCh {
		var out T
		if err := json.Unmarshal(ele.Body, &out); err != nil {
			return err
		}
		ack := handler(out)
		switch ack {
		case Ack:
			ele.Ack(false)
			fmt.Print("Acknowlegment")
		case NackRequeue:
			ele.Nack(false, true)
			fmt.Print("Nack requeue")
		case NackDiscard:
			ele.Nack(false, false)
			fmt.Print("Nack dicard")
		}
	}
	return nil
}
