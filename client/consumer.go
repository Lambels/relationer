package client

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
)

type consumer struct {
	*amqp.Connection
	*amqp.Channel

	private  bool
	isClosed atomic.Value // used to differentiate between failed connection or forced shutdown to stop redialing.
	done     chan struct{}
	mu       sync.Mutex
	index    uint
	idCount  bool
	notify   map[uint]chan<- *Message
}

// TODO: add conf struct
func newConsumer(URL string, bindings []string, reconnect bool) (*consumer, error) {
	cons := &consumer{
		done:   make(chan struct{}),
		notify: make(map[uint]chan<- *Message),
	}
	cons.isClosed.Store(false)

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

	del, err := cons.Channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}

	if reconnect {
		go cons.redial()
	}
	go cons.handle(del)

	return cons, nil
}

func (c *consumer) redial(every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		<-ticker.C
		if c.isClosed.Load().(bool) {
			return
		}
		select {
		case <-c.done:

		}
	}
}

func (c *consumer) attachRecv(chan<- *Message) {

}

func (c *consumer) removeRecv(chan<- *Message) {

}

func (c *consumer) handle(del <-chan amqp.Delivery) {
	for d := range del {
		msg := &Message{
			Type: d.Type,
			Data: d.Body,
		}
		c.share(msg)
	}
	c.done <- struct{}{}
}

func (c *consumer) share(msg *Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, c := range c.notify {
		c <- msg
	}
}

func (c *consumer) shutdown() error {
	if c.Connection == nil {
		return nil
	}
	c.isClosed.Store(true)

	// channel.Cancel
	return nil
}
