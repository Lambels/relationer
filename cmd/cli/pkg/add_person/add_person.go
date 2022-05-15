package addperson

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Lambels/relationer/cmd/cli/pkg/root"
	"github.com/Lambels/relationer/internal"
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
		Name:       "add-person",
		ShortUsage: "relationer add-person",
		ShortHelp:  "Create a user (node)",
		Exec:       cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return errors.New("add-person requires 1 argument")
	}

	name := args[0]
	user := &internal.Person{
		Name: name,
	}
	if err := c.rootConfig.Client.AddPerson(ctx, user); err != nil {
		return err
	}

	if c.rootConfig.Verbose {
		fmt.Fprintf(c.out, "created person with id %q at %v OK\n", user.ID, user.CreatedAt)
	}

	return nil
}
