package getdepth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/Lambels/relationer/cmd/cli/pkg/root"
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
		Name:       "get-depth",
		ShortUsage: "relationer get-depth",
		ShortHelp:  "Get depth between 2 nodes",
		Exec:       cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return errors.New("get-depth requires 2 argument")
	}
	id1, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("non int argument")
	}
	id2, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.New("non int argument")
	}

	start := time.Now()
	depth, err := c.rootConfig.Client.GetDepth(ctx, int64(id1), int64(id2))
	if err != nil {
		return err
	}
	fmt.Fprintf(c.out, "%v (endpoints included)\n", depth)

	if c.rootConfig.Verbose {
		fmt.Fprintf(c.out, "OK\n")
		fmt.Fprintf(c.out, "Process took %v \n", time.Since(start))

	}

	return nil
}
