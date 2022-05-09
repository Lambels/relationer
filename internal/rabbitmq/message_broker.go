package rabbitmq

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/Lambels/relationer/internal"
	"github.com/streadway/amqp"
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
	return s.pushMsg(ctx, "person.created", person)
}

func (s *RabbitMq) CreatedFriendship(ctx context.Context, friendship internal.Friendship) error {
	return s.pushMsg(ctx, "friendship.created", friendship)
}

func (s *RabbitMq) DeletedPerson(ctx context.Context, person *internal.Person) error {
	return s.pushMsg(ctx, "person.deleted", person)
}

func (s *RabbitMq) pushMsg(ctx context.Context, routingKey string, val interface{}) error {
	var buf *bytes.Buffer
	if err := json.NewEncoder(buf).Encode(val); err != nil {
		return internal.WrapError(err, internal.EINTERNAL, "json.Encode")
	}

	if err := s.ch.Publish(
		"relations",
		routingKey,
		false,
		false,
		amqp.Publishing{
			AppId:       "rest-server",
			ContentType: "application/json",
			Body:        buf.Bytes(),
			Timestamp:   time.Now(),
		},
	); err != nil {
		return internal.WrapError(err, internal.EINTERNAL, "ch.Publish")
	}
	return nil
}
