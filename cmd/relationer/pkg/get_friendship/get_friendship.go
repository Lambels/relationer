package getfriendship

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/Lambels/relationer/cmd/relationer/pkg/root"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
	rootConfig *root.Config
	out        io.Writer
}

func New(rootConfig *root.Config, out io.Writer) *ffcli.Command {
	cfg := Config{
		rootConfig: rootConfig,
		out:        out,
	}

	return &ffcli.Command{
		Name:       "get-friendship",
		ShortUsage: "relationer get-friendship",
		ShortHelp:  "Get the relationships of a person",
		Exec:       cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return errors.New("get-friends requires 1 argument")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("non int argument")
	}

	start := time.Now()
	friendship, err := c.rootConfig.Client.GetFriendship(ctx, int64(id))
	if err != nil {
		return err
	}
	fmt.Fprintf(c.out, "%v is friends with:\n", friendship.P1.Name)
	for _, friend := range friendship.With {
		fmt.Fprintf(c.out, "  ↪%v\n", friend)
	}

	if c.rootConfig.Verbose {
		fmt.Fprintf(c.out, "OK\n")
		fmt.Fprintf(c.out, "Process took %v \n", time.Since(start))
	}

	return nil
}
