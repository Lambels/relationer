package removeperson

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
		Name:       "remove-person",
		ShortUsage: "relationer remove-person",
		ShortHelp:  "Delete a user (node)",
		Exec:       cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return errors.New("remove-person requires 1 argument")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("non int argument")
	}

	start := time.Now()
	if err := c.rootConfig.Client.RemovePerson(ctx, int64(id)); err != nil {
		return err
	}
	if c.rootConfig.Verbose {
		fmt.Fprintf(c.out, "removed person with id: %v OK\n", id)
		fmt.Fprintf(c.out, "Process took %v \n", time.Since(start))
	}

	return nil
}
