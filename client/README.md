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

# More Examples

## Client Listening:
```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/Lambels/relationer/client"
)

func main() {
	c := client.NewClient(nil)

    // start a new detached listener.
	recv, err := c.ListenDetached(context.Background())
	if err != nil {
		log.Fatal(err)
	}

    // close will make sure that it will tear any connections and close any consumers left behind.
	defer c.Close()

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