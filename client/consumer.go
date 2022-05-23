package client

import (
	"sync"

	"github.com/streadway/amqp"
)

type consumer struct {
	*amqp.Connection
	*amqp.Channel

	done   chan struct{}
	mu     sync.Mutex
	notify []chan *Message // own custom type
}

func newConsumer(URL string, bindings []string) (*consumer, error) {
	cons := &consumer{
		done:   make(chan struct{}),
		notify: make([]chan *Message, 0),
	}

	var err error
	cons.Connection, err = amqp.Dial(URL)
	if err != nil {
		return nil, err
	}

	cons.Channel, err = cons.Connection.Channel()
	if err != nil {
		return nil, err
	}

	if err = cons.Channel.ExchangeDeclarePassive(
		"relationer", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,
	); err != nil {
		return nil, err
	}

	q, err := cons.Channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	// bind ...
	for _, bind := range bindings {
		if err := cons.Channel.QueueBind(
			q.Name,       // queue name
			bind,         // routing key
			"relationer", // exchange
			false,
			nil,
		); err != nil {
			return nil, err
		}
	}

	return cons, nil
}

func (c consumer) shutdown() error {
	if c.Connection == nil {
		return nil
	}

	// channel.Cancel
	return nil
}
