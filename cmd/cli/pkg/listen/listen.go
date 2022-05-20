package listen

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Lambels/relationer/cmd/cli/pkg/root"
	"github.com/Lambels/relationer/internal"
	"github.com/Lambels/relationer/internal/rabbitmq"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/streadway/amqp"
)

type Config struct {
	rootConfig *root.Config
	out        io.Writer
	all        bool
}

func New(rootConfig *root.Config, out io.Writer) *ffcli.Command {
	cfg := Config{
		rootConfig: rootConfig,
		out:        out,
	}

	fs := flag.NewFlagSet("relationer listen", flag.ExitOnError)
	fs.BoolVar(&cfg.all, "all", false, "listen to all events")
	rootConfig.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "listen",
		ShortUsage: "relationer listen",
		ShortHelp:  "Listen for events",
		FlagSet:    fs,
		Exec:       cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) == 0 && !c.all {
		return errors.New("get-person requires a list of routing keys, to listen to all events run with --all flag set")
	}

	for ses := range c.redial(ctx, "amqp://guest:guest@localhost:5672") {
		q, err := ses.QueueDeclare(
			"",    // name
			false, // durable
			false, // delete when unused
			true,  // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("ses.QueueDeclare: %w", err)
		}

		for _, s := range args {
			if err := ses.Channel.QueueBind(
				q.Name,       // queue name
				s,            // routing key
				"relationer", // exchange
				false,
				nil,
			); err != nil {
				return fmt.Errorf("ch.QueueBind: %w", err)
			}
		}

		if c.all {
			if err := ses.Channel.QueueBind(
				q.Name,       // queue name
				"#",          // routing key
				"relationer", // exchange
				false,
				nil,
			); err != nil {
				return fmt.Errorf("ch.QueueBind: %w", err)
			}
		}

		deliveries, err := ses.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto ack
			false,  // exclusive
			false,  // no local
			false,  // no wait
			nil,    // args
		)
		if err != nil {
			return fmt.Errorf("ses.Consume: %w", err)
		}

		if c.rootConfig.Verbose {
			fmt.Fprintln(c.out, "subscribed and ready to recieve...")
		}

		for msg := range deliveries {
			switch msg.Type {
			case rabbitmq.MesssagePersonCreated:
				var person internal.Person
				if err := json.Unmarshal(msg.Body, &person); err != nil {
					return fmt.Errorf("failed to unmarshal message body")
				}
				fmt.Fprintf(c.out, "[New Person] Id: %v | Name: %v | Created At: %v\n", person.ID, person.Name, person.CreatedAt)

			case rabbitmq.MessageFriendshipCreated:
				var friendship internal.Friendship
				if err := json.Unmarshal(msg.Body, &friendship); err != nil {
					return fmt.Errorf("failed to unmarshal message body")
				}
				fmt.Fprintf(c.out, "[New Friendship] Person 1: %v with Person 2: %v\n", friendship.P1.ID, friendship.With[0])

			case rabbitmq.MessagePersonDeleted:
				var payload map[string]int64
				if err := json.Unmarshal(msg.Body, &payload); err != nil {
					return fmt.Errorf("failed to unmarshal message body")
				}
				fmt.Fprintf(c.out, "[Removed Person] Id: %v\n", payload["id"])

			default:
				return fmt.Errorf("message type unknown")
			}
		}
	}

	return nil
}

// session bundels the amqp connection with the channel.
type session struct {
	*amqp.Connection
	*amqp.Channel
}

// Close closses the connection and channel.
func (s session) Close() error {
	if s.Connection == nil {
		return nil
	}

	return s.Connection.Close()
}

// redial continously tries to redial URL.
func (c *Config) redial(ctx context.Context, URL string) <-chan session {
	comm := make(chan session, 1)

	go func() {
		for {
			conn, err := amqp.Dial(URL)
			if err != nil {
				fmt.Fprintf(c.out, "amqp.Dial: %v", err)
				os.Exit(1)
			}

			ch, err := conn.Channel()
			if err != nil {
				fmt.Fprintf(c.out, "conn.Channel: %v", err)
				os.Exit(1)
			}

			if err := ch.ExchangeDeclare(
				"relationer", // name
				"topic",      // type
				true,         // durable
				false,        // auto-deleted
				false,        // internal
				false,        // no-wait
				nil,          // args
			); err != nil {
				fmt.Fprintf(c.out, "ch.ExchangeDeclare: %v", err)
				os.Exit(1)
			}

			select {
			case comm <- session{conn, ch}:
			case <-ctx.Done():
				if c.rootConfig.Verbose {
					fmt.Fprint(c.out, "shutting down session factory")
					return
				}
			}
		}
	}()

	return comm
}
