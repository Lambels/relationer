package client

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
)

const DefaultBrokerURL = "amqp://guest:guest@localhost:5672"

// ConsumerConfig is used for the creation of consumers, each client has one consumer config
// which is used to create all consumers via the: StartListenDetached and StartListenAttached
// methods.
type ConsumerConfig struct {
	// The URL of the broker (rabbitmq) - default: amqp://guest:guest@localhost:5672
	URL string
	// Binding keys to the exchange - defualt: # (all)
	BindingKeys []string

	// Used as timeout to check the health of current connection. (reconnect if needed or close conn)
	Pulse time.Duration
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
	idCount  int
	notify   map[int]chan<- *Message
}

func newConsumer(conf *ConsumerConfig) (*consumer, error) {
	cons := &consumer{
		done:   make(chan struct{}, 1),
		notify: make(map[int]chan<- *Message),
	}
	cons.isClosed.Store(false)

	// set consumer configuration.
	if conf == nil {
		cons.conf = &ConsumerConfig{
			URL:         DefaultBrokerURL,
			BindingKeys: []string{"#"},
			Pulse:       0,
		}
	} else {
		cons.conf = conf
	}

	// establish the amqp connection + channel.
	if err := cons.establishConnection(); err != nil {
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

	if cons.conf.Pulse != 0 {
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
		// always check if c.isClosed flag set and steal the done value.
		if c.isClosed.Load().(bool) {
			return
		}
		select {
		case <-c.done: // past c.handle() exited
			if err := c.establishConnection(); err != nil {
				c.isClosed.Store(true)
				c.closeRecievers()
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
				c.isClosed.Store(true)
				c.closeRecievers()
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
					c.isClosed.Store(true)
					c.closeRecievers()
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
				c.isClosed.Store(true)
				c.closeRecievers()
				return
			}

			// start new handler.
			go c.handle(del)
		default:
		}
	}
}

func (c *consumer) attachRecv(recv chan<- *Message) (int, error) {
	// consumer dead, signal to remove from consumers slice.
	if c.isClosed.Load().(bool) {
		return -1, fmt.Errorf("consumer closed")
	}

	c.mu.Lock()
	id := c.idCount
	c.notify[id] = recv
	c.idCount++
	c.mu.Unlock()
	return id, nil
}

func (c *consumer) removeRecv(id int, isRoot bool) error {
	c.mu.Lock()
	if isRoot || len(c.notify) == 0 {
		c.mu.Unlock()
		c.shutdown()
		return fmt.Errorf("consumer empty")
	}
	delete(c.notify, id)
	c.mu.Unlock()
	return nil
}

func (c *consumer) handle(del <-chan amqp.Delivery) {
	for d := range del {
		msg := &Message{
			Type: d.Type,
			Data: d.Body,
		}
		c.share(msg)
	}
	if c.conf.Pulse == 0 && !c.isClosed.Load().(bool) { // the connection is closed and cant be recovered.
		// connection already closed.
		// no need to go through shutdown method.
		c.closeRecievers()
		c.isClosed.Store(true)
	}

	// no need for select, buffered channel.
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
	if c.Connection == nil || c.isClosed.Load().(bool) {
		return nil
	}
	c.isClosed.Store(true)
	defer c.closeRecievers()

	if err := c.Connection.Close(); err != nil {
		return err
	}

	<-c.done

	return nil
}

func (c *consumer) closeRecievers() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, recv := range c.notify {
		close(recv)
	}
}

func (c *consumer) establishConnection() error {
	var err error
	c.Connection, err = amqp.Dial(c.conf.URL)
	if err != nil {
		return err
	}

	c.Channel, err = c.Connection.Channel()
	if err != nil {
		return err
	}

	if err = c.Channel.ExchangeDeclarePassive(
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
