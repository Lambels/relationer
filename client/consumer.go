package client

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
)

type ConsumerConfig struct {
	URL         string
	BindingKeys []string

	Reconnect bool
	Pulse     time.Duration // pulse should only be set if reconnect is true.
}

// consumer represents a TCP connection to a message broker.
// a consumer may have multiple recievers attached to it, yet not recommended
// if a TCP connection gets killed you may want your other recievers un-harmed.
type consumer struct {
	*amqp.Connection
	*amqp.Channel

	isClosed atomic.Value // used to differentiate between failed connection or forced shutdown to stop redialing.
	done     chan struct{}
	conf     *ConsumerConfig
	mu       sync.Mutex
	index    int
	idCount  int
	notify   map[int]chan<- *Message
}

func newConsumer(conf *ConsumerConfig) (*consumer, error) {
	cons := &consumer{
		done:   make(chan struct{}),
		notify: make(map[int]chan<- *Message),
	}
	cons.isClosed.Store(false)

	// set consumer configuration.
	if conf == nil {
		cons.conf = &ConsumerConfig{
			URL:         "amqp://guest:guest@localhost:5672",
			BindingKeys: []string{"#"},
		}
	} else {
		cons.conf = conf
	}

	// establish the amqp connection + channel.
	if err := establishConnection(cons); err != nil {
		return nil, err
	}

	// create a new random name queue.
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

	// bind queue.
	for _, bind := range cons.conf.BindingKeys {
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

	// start consuming.
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

	if cons.conf.Reconnect {
		go cons.redial()
	}
	go cons.handle(del)

	return cons, nil
}

func (c *consumer) redial() {
	ticker := time.NewTicker(c.conf.Pulse)
	defer ticker.Stop()
	for {
		<-ticker.C
		// always check if c.isClosed flag set and not consumer the done channel.
		if c.isClosed.Load().(bool) {
			return
		}
		select {
		case <-c.done: // past c.handle() exited
			if err := establishConnection(c); err != nil {
				// TODO: figure out how to signal failure
				return
			}

			// create a new random name queue.
			q, err := c.Channel.QueueDeclare(
				"",    // name
				false, // durable
				false, // delete when unused
				true,  // exclusive
				false, // no-wait
				nil,   // arguments
			)
			if err != nil {
				// TODO: figure out how to signal failure
				return
			}

			// bind queue.
			for _, bind := range c.conf.BindingKeys {
				if err := c.Channel.QueueBind(
					q.Name,       // queue name
					bind,         // routing key
					"relationer", // exchange
					false,
					nil,
				); err != nil {
					// TODO: figure out how to signal failure
					return
				}
			}

			// start consuming.
			del, err := c.Channel.Consume(
				q.Name, // queue
				"",     // consumer
				true,   // auto ack
				false,  // exclusive
				false,  // no local
				false,  // no wait
				nil,    // args
			)
			if err != nil {
				// TODO: figure out how to signal failure
				return
			}

			// start new handler.
			go c.handle(del)
		default:
		}
	}
}

// TODO: maybe return error to signal removal from slice in client.
// TODO: error field? to track any redial error.
func (c *consumer) attachRecv(recv chan<- *Message) int {
	c.mu.Lock()
	id := c.idCount
	c.notify[id] = recv
	c.idCount++
	c.mu.Unlock()
	return id
}

func (c *consumer) removeRecv(id int, isRoot bool) {
	c.mu.Lock()
	delete(c.notify, id)
	if isRoot || len(c.notify) == 0 {
		c.mu.Unlock()
		c.shutdown()
		return
	}
	c.mu.Unlock()

}

func (c *consumer) handle(del <-chan amqp.Delivery) {
	for d := range del {
		msg := &Message{
			Type: d.Type,
			Data: d.Body,
		}
		c.share(msg)
	}
	c.done <- struct{}{} // TODO: do something with this signal
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

func establishConnection(cons *consumer) error {
	var err error
	cons.Connection, err = amqp.Dial(cons.conf.URL)
	if err != nil {
		return err
	}

	cons.Channel, err = cons.Connection.Channel()
	if err != nil {
		return err
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
		return err
	}

	return nil
}
