package client

import (
	"context"
	"net/http"
	"sync"
	"time"

	rClient "github.com/Lambels/relationer/internal/client"
	"github.com/Lambels/relationer/internal/service"
)

const (
	DefaultTimeout = time.Second * 5
	DefaultURL     = "http://localhost:8080"
)

type Client struct {
	client service.GraphStore

	bindings []string

	mu        sync.Mutex // protects bottom fields
	consumers []*consumer
	lastx     uint
}

type ClientConfig struct {
	URL      string
	Timeout  time.Duration
	Bindings []string
}

func NewClient(conf *ClientConfig) *Client {
	c := &Client{
		consumers: make([]*consumer, 0),
		bindings:  make([]string, 0),
	}
	if conf == nil {
		c.client = rClient.NewClient(http.Client{Timeout: DefaultTimeout}, DefaultURL)
		c.bindings = append(c.bindings, "#")
	} else {
		c.client = rClient.NewClient(http.Client{Timeout: conf.Timeout}, conf.URL)
		c.bindings = append(c.bindings, conf.Bindings...)
	}
	return c
}

func (c *Client) ListenDetached(ctx context.Context) <-chan *Message {
	return c.listen(ctx, true)
}

func (c *Client) ListenAttached(ctx context.Context) <-chan *Message {
	return c.listen(ctx, false)
}

func (c *Client) listen(ctx context.Context, independant bool) <-chan *Message {

}
