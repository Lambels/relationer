package client

//TODO: Add more examples in readme.
import (
	"context"
	"fmt"
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

// Client is the interface between the user and the relationer server.
type Client struct {
	// REST client implementation.
	client service.GraphStore

	// The consumer factory config.
	consConf *ConsumerConfig

	mu           sync.Mutex // protects bottom fields.
	consumerPool []*consumer
	lastx        int
}

// ClientConfig represents the configurable fields.
type ClientConfig struct {
	// The relationer server URL - default: http://localhost:8080
	URL string
	// The http client used by the REST client - defualt: http client with 5 second timeout
	Client *http.Client

	// The consumer factory configuration used by the client to generate all consumers.
	ConsumerConfig *ConsumerConfig
}

// New creates a new relationer client with the provided config.
//
// To use the default config run
//	client.New(nil)
func New(conf *ClientConfig) *Client {
	c := &Client{
		consumerPool: make([]*consumer, 0),
	}
	if conf == nil {
		c.client = rClient.NewClient(&http.Client{Timeout: DefaultTimeout}, DefaultURL)
	} else {
		c.client = rClient.NewClient(conf.Client, conf.URL)
		c.consConf = conf.ConsumerConfig
	}
	return c
}

// Close tears any possible amqp connections and consumers taking with it all the registered
// channels (recievers).
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

// TODO: parse errors // unwrap from internal client.
// REST -------------------------------------------------------------------------------------

// GetDepth gets the depth between two nodes (including the endpoint nodes).
func (c *Client) GetDepth(ctx context.Context, id1, id2 int64) (int, error) {
	return c.client.GetDepth(ctx, id1, id2)
}

// GetFriendship gets the friendships (relationships) the person with id: id has.
func (c *Client) GetFriendship(ctx context.Context, id int64) (internal.Friendship, error) {
	return c.client.GetFriendship(ctx, id)
}

// GetPerson fetches the person with id: id.
func (c *Client) GetPerson(ctx context.Context, id int64) (*internal.Person, error) {
	return c.client.GetPerson(ctx, id)
}

// GetAll returns the graph of the current state.
func (c *Client) GetAll(ctx context.Context) ([]internal.Friendship, error) {
	return c.client.GetAll(ctx)
}

// AddFriendship creates a new friendship (one-way) between p1 and p2.
func (c *Client) AddFriendship(ctx context.Context, p1, p2 int64) error {
	f := internal.Friendship{
		P1:   &internal.Person{ID: p1},
		With: []int64{p2},
	}
	return c.client.AddFriendship(ctx, f)
}

// AddPerson adds person with name: name and returns the id if successful.
func (c *Client) AddPerson(ctx context.Context, name string) (int64, error) {
	p := internal.Person{
		Name: name,
	}
	if err := c.client.AddPerson(ctx, &p); err != nil {
		return -1, err
	}

	return p.ID, nil
}

// RemovePerson removes person with id: id.
func (c *Client) RemovePerson(ctx context.Context, id int64) error {
	return c.client.RemovePerson(ctx, id)
}

// Listen -----------------------------------------------------------------------------------

// StartListenDetached will start a separate connection to the message-broker (rabbitmq)
// and subscribe the recieving channel to it.
//
// canceling the context provided to this function call will destroy the whole rabbitmq connection taking
// away with it the initial subscribed channel and any other attached channels with
//	StartListenAttached()
// this way of listening to the messages is recommanded as separate connections for each
// recieving channel will make recievers more safe and independant.
func (c *Client) StartListenDetached(ctx context.Context) (<-chan *Message, error) {
	return c.listen(ctx, true)
}

// StartListenAttached will attach to an existing connection if possible, only scenario where
// it will create a new connection is when the connection pool is empty.
//
// Canceling the context will end up in one of these cases:
//
// 1) When a new connection is created by StartListenAttached cancelling the context passed
// to the function call will destroy the connection only when there are no more recievers
// attached to it.
//
// 2) When the reciever is attached to an existing connection cancelling the context will just
// remove the reciever without killing the connection, the only way that connection can be killed
// is if the root reciever (reciever created with StartListenDetached) is killed or an error happens
// in the connection.
//
// Grouping - group multiple recievers to one root reciever (controlled):
//
// call (*Client).StartListenDetached(ctx) followed by synchronised
// (*Client).StartListenAttached(ctx, true) , these followed calls will skip the round-robin like
// load balancer algorithm and attach all the recievers to the last consumer (created by (*Client).StartListenDetached(ctx))
// giving you grouped consumer.
//
// Attention: A recv attached to a connection can be closed at any time! You can bundle up
// recievers with syncronised calls to the client making one root reciever and many dependant
// recievers.
func (c *Client) StartListenAttached(ctx context.Context, last bool) (<-chan *Message, error) {
	if last {
		return c.listenLast(ctx)
	}
	return c.listen(ctx, false)
}

func (c *Client) listenLast(ctx context.Context) (<-chan *Message, error) {
	c.mu.Lock()
	if len(c.consumerPool) == 0 {
		c.mu.Unlock()
		return nil, fmt.Errorf("no active consumers.")
	}

	recv := make(chan *Message)
	cons := c.consumerPool[len(c.consumerPool)-1]
	id, err := cons.attachRecv(recv)
	if err != nil { // dead?
		c.consumerPool = c.consumerPool[:len(c.consumerPool)-1]
		return nil, err
	}
	c.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
		}

		if err := cons.removeRecv(id, false); err != nil { // return early, close channel 2 times will cause a panic.
			return
		}
		close(recv) // close chan.
	}()

	return recv, nil
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

		if err := cons.removeRecv(id, root); err != nil { // return early, close channel 2 times will cause a panic.
			return
		}
		close(recv) // close chan.
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
			c.lastx++
		}
	}

	id, err := cons.attachRecv(recv)
	if err != nil { // dead?
		c.consumerPool = append(c.consumerPool[:c.lastx], c.consumerPool[c.lastx+1:]...) // remove.
		c.asignConsumerReciever(recv, root)
	}
	return cons, id, nil
}
