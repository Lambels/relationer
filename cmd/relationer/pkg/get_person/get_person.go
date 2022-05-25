package getperson

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
		Name:       "get-person",
		ShortUsage: "relationer get-person",
		ShortHelp:  "Get the person with provided id",
		Exec:       cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return errors.New("get-person requires 1 argument")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("non int argument")
	}

	start := time.Now()
	person, err := c.rootConfig.Client.GetPerson(ctx, int64(id))
	if err != nil {
		return err
	}
	fmt.Fprintf(c.out, "Id: %v | Name: %v | Created At: %v\n", person.ID, person.Name, person.CreatedAt)

	if c.rootConfig.Verbose {
		fmt.Fprintf(c.out, "OK\n")
		fmt.Fprintf(c.out, "Process took %v \n", time.Since(start))
	}

	return nil
}
