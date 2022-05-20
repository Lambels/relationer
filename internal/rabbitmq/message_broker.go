package rabbitmq

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/Lambels/relationer/internal"
	"github.com/streadway/amqp"
)

// types for: https://www.rabbitmq.com/publishers.html#message-properties
const (
	MesssagePersonCreated    = "person.created"
	MessagePersonDeleted     = "person.deleted"
	MessageFriendshipCreated = "friendship.created"
)

type RabbitMq struct {
	ch *amqp.Channel
}

func NewRabbitMq(ch *amqp.Channel) *RabbitMq {
	return &RabbitMq{
		ch: ch,
	}
}

func (s *RabbitMq) CreatedPerson(ctx context.Context, person *internal.Person) error {
	return s.pushMsg(ctx, "person.created", person, MesssagePersonCreated)
}

func (s *RabbitMq) CreatedFriendship(ctx context.Context, friendship internal.Friendship) error {
	return s.pushMsg(ctx, "friendship.created", friendship, MessageFriendshipCreated)
}

func (s *RabbitMq) DeletedPerson(ctx context.Context, id int64) error {
	return s.pushMsg(ctx, "person.deleted", map[string]int64{"id": id}, MessagePersonDeleted)
}

func (s *RabbitMq) pushMsg(ctx context.Context, routingKey string, val interface{}, t string) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(val); err != nil {
		return internal.WrapError(err, internal.EINTERNAL, "json.Encode")
	}

	if err := s.ch.Publish(
		"relationer",
		routingKey,
		false,
		false,
		amqp.Publishing{
			AppId:           "rest-server",
			ContentEncoding: "application/json",
			Type:            t,
			Body:            buf.Bytes(),
			Timestamp:       time.Now(),
		},
	); err != nil {
		return internal.WrapError(err, internal.EINTERNAL, "ch.Publish")
	}
	return nil
}
