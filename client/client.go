package client

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Lambels/relationer/internal"
	rClient "github.com/Lambels/relationer/internal/client"
	"github.com/Lambels/relationer/internal/service"
)

const (
	DefaultTimeout = time.Second * 5
	DefaultURL     = "http://localhost:8080"
)

type Client struct {
	client service.GraphStore

	consConf *ConsumerConfig

	mu           sync.Mutex // protects bottom fields
	consumerPool []*consumer
	lastx        int
}

type ClientConfig struct {
	URL     string
	Timeout time.Duration

	ConsConf *ConsumerConfig
}

func NewClient(conf *ClientConfig) *Client {
	c := &Client{
		consumerPool: make([]*consumer, 0),
	}
	if conf == nil {
		c.client = rClient.NewClient(http.Client{Timeout: DefaultTimeout}, DefaultURL)
	} else {
		c.client = rClient.NewClient(http.Client{Timeout: conf.Timeout}, conf.URL)
		c.consConf = conf.ConsConf
	}
	return c
}

// Close tears any possible amqp connections and consumers, also closes any reciever channels.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cons := range c.consumerPool {
		if err := cons.shutdown(); err != nil {
			return err
		}
	}

	return nil
}

// REST -------------------------------------------------------------------------------------

func (c *Client) GetDepth(ctx context.Context, id1, id2 int64) (int, error) {
	return c.client.GetDepth(ctx, id1, id2)
}

func (c *Client) GetFriendship(ctx context.Context, id int64) (internal.Friendship, error) {
	return c.client.GetFriendship(ctx, id)
}

func (c *Client) GetPerson(ctx context.Context, id int64) (*internal.Person, error) {
	return c.client.GetPerson(ctx, id)
}

func (c *Client) GetAll(ctx context.Context) ([]internal.Friendship, error) {
	return c.client.GetAll(ctx)
}

func (c *Client) AddFriendship(ctx context.Context, f internal.Friendship) error {
	return c.client.AddFriendship(ctx, f)
}

func (c *Client) AddPerson(ctx context.Context, p *internal.Person) error {
	return c.client.AddPerson(ctx, p)
}

func (c *Client) RemovePerson(ctx context.Context, id int64) error {
	return c.client.RemovePerson(ctx, id)
}

// Listen -----------------------------------------------------------------------------------

func (c *Client) ListenDetached(ctx context.Context) (<-chan *Message, error) {
	return c.listen(ctx, true)
}

func (c *Client) ListenAttached(ctx context.Context) (<-chan *Message, error) {
	return c.listen(ctx, false)
}

func (c *Client) listen(ctx context.Context, root bool) (<-chan *Message, error) {
	recv := make(chan *Message)

	c.mu.Lock()
	cons, id, err := c.asignConsumerReciever(recv, root)
	if err != nil {
		return nil, err
	}
	c.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
		}

		cons.removeRecv(id, root)
		close(recv) // close chan
	}()

	return recv, nil
}

func (c *Client) asignConsumerReciever(recv chan *Message, root bool) (*consumer, int, error) {
	var cons *consumer
	if root { // create separate connection (consumer) for root reciever.
		var err error
		cons, err = newConsumer(c.consConf)
		if err != nil {
			return nil, -1, err
		}

		c.consumerPool = append(c.consumerPool, cons)
	} else {
		if len(c.consumerPool) == 0 { // no current consumers.
			// create the first consumer and just attach the reciever to it.
			var err error
			cons, err = newConsumer(c.consConf)
			if err != nil {
				return nil, -1, err
			}

			c.consumerPool = append(c.consumerPool, cons)
		} else { // add reciever to existing consumer.
			if c.lastx >= len(c.consumerPool) { // if index over slice length start from the begining.
				c.lastx = 0
			}

			cons = c.consumerPool[c.lastx] // get attaching consumer
			c.lastx++                      // TODO: change pointer only when 100 % sure that recv attached
		}
	}

	id, err := cons.attachRecv(recv)
	if err != nil { // dead?
		c.consumerPool = append(c.consumerPool[:c.lastx], c.consumerPool[c.lastx+1:]...) // remove.
		c.asignConsumerReciever(recv, root)
	}
	return cons, id, nil
}
