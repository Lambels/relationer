# Relationer golang client
Use the golang realtioner client to interact programatically with the relationer server. Read messages and preform CRUD opperations.

# Getting started

Go get it:
```
go get github.com/Lambels/relationer
```

## Simple example:
In this example you will create 2 people and link them together mutually.
```go
package main

import (
    "log"
    "context"

    "github.com/Lambels/relationer/client"
)

func main() {
    rClient := client.New(nil) // default client settings.
    
    bgCtx := context.Background()

    // create first person with name: 'Lambels'
    id, err := rClient.AddPerson(bgCtx, "Lambels")
    if err != nil {
        log.Fatal(err)
    }

    // create second person with name: 'Your-Name'
    id2, err := rClient.AddPerson(bgCtx, "Your-Name")
    if err != nil {
        log.Fatal(err)
    }

    // link them together (both ways, ie: mutual friendship)
    if err := rClient.AddFriendship(bgCtx, id1, id2); err != nil {
        log.Fatal(err)
    }
    if err := rClient.AddFriendship(bgCtx, id2, id1); err != nil {
        log.Fatal(err)
    }
}
```

## Client Settings:
When initiating a client you can pass a configuration struct to the initializator to customize the behaviour of the client.

The client configuration can be split into 2 parts: `client.ClientConfig` and `client.ConsumerConfig`, the `client.ClientConfig`, configures the behaviour of `REST` methods such as `client.Get*` or `client.Add*` while `client.ConsumerConfig` configures the behaviour of `client.Start*` methods.

### ClientConfig:
```go
// to use the default value for the client cofig pass nil.
client.New(&client.ClientConfig{
    URL: "http://localhost:3000", // default value - http://localhost:8080
    Client: &http.Client{Timeout: 5 * time.Second}, // default value - http.DefaultClient
    ConsumerConfig: nil, // use default consumer config
})
```

### ConsumerConfig:
When in reconnecting state separate ticking go routine which sleeps in intervals of `Pulse` checks the health of the consumer, if the connection is closed it will attempt a redial or close the consumer if the redial fails.
```go
// to use the default value for the consumer config pass nil to ConsumerConfig.
client.New(&client.ClientConfig{
    URL: "http://localhost:3000", // default value - http://localhost:8080
    Client: &http.Client{Timeout: 5 * time.Second}, // default value - http.DefaultClient
    ConsumerConfig: &client.ConsumerConfig{
        URL: "amqp://guest:guest@localhost:8080" // default value - amqp://guest:guest@localhost:5672
        BindingKeys: []string{"person.created", "person.deleted"} // default value - "#" (all messages)
        Reconnect: true, // default value - false
        Pulse: 5 * time.Second, // default value - 0
    },
})
```

# More Examples

## Client Listening:
```go
package main

import (
	"context"
	"log"

	"github.com/Lambels/relationer/client"
)

func main() {
	c := client.New(nil)
    
    // close will make sure that it will tear any connections and close any consumers left behind.
	defer c.Close()
    
    // start a new detached listener.
	recv, err := c.StartListenDetached(context.Background())
	if err != nil {
		log.Fatal(err)
	}
    
    // loop over messages.
	go func() { 
        log.Println("Started listening...")
		for msg := range recv1 {
			log.Printf("[%v]: %v\n", msg.Type, msg.Data)
		}
	}()
    
    log.Println("Ctrl-C to stop listening.")
	select {}
}
```

## Client listening (attached):
```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/Lambels/relationer/client"
)

func logMessages(recv <-chan *client.Message) {
    go func() {
        log.Println("Started listening...")
		for msg := range recv1 {
			log.Printf("[%v]: %v\n", msg.Type, msg.Data)
		}
    }()
}

func main() {
	c := client.New(nil)
    
    // close will make sure that it will tear any connections and close any consumers left behind.
	defer c.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()

    // start a new detached listener with a timeout context.
	recv, err := c.StartListenDetached(ctx)
	if err != nil {
		log.Fatal(err)
	}
    logMessages(recv)
    
    for i := 0; i < 4; i++ {    
        // start a new reciever without creating a new connection to the rabbitmq server.
        // the true parameter will attach the reciever strictly to the last connection added
        // and skip the load balancing algorithm.
        recv, err := c.StartListenAttached(context.Background(), true)
	    if err != nil {
		    log.Fatal(err)
	    }
        logMessages(recv)
    }
        
    // the resulting connections to the rabbitmq server will be one. (detached listener)
    log.Println("Ctrl-C to stop listening.")
	select {}
}
```