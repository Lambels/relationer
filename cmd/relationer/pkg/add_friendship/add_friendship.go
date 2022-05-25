package addfriendship

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/Lambels/relationer/cmd/cli/pkg/root"
	"github.com/Lambels/relationer/internal"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
	rootConfig *root.Config
	out        io.Writer
	mutual     bool
}

func New(rootConfig *root.Config, out io.Writer) *ffcli.Command {
	cfg := Config{
		rootConfig: rootConfig,
		out:        out,
	}

	fs := flag.NewFlagSet("relationer add-friendship", flag.ExitOnError)
	fs.BoolVar(&cfg.mutual, "mutual", false, "create a bi-directional edge")

	return &ffcli.Command{
		Name:       "add-friendship",
		ShortUsage: "relationer add-friendship",
		ShortHelp:  "Create a friendship (edge) uni-directional from id1 -> id2",
		FlagSet:    fs,
		Exec:       cfg.Exec,
	}
}

func (c *Config) Exec(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return errors.New("add-friendship requires 2 argument")
	}
	id1, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("non int argument")
	}
	id2, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.New("non int argument")
	}

	friendship := internal.Friendship{
		P1:   &internal.Person{ID: int64(id1)},
		With: []int64{int64(id2)},
	}
	start := time.Now()
	if err := c.rootConfig.Client.AddFriendship(ctx, friendship); err != nil {
		return err
	}
	if c.mutual {
		p1 := friendship.P1
		friendship.P1 = &internal.Person{ID: friendship.With[0]}
		friendship.With[0] = p1.ID
		if err := c.rootConfig.Client.AddFriendship(ctx, friendship); err != nil {
			return err
		}
	}

	if c.rootConfig.Verbose {
		fmt.Fprintf(c.out, "created friendship between %v and %v OK\n", id1, id2)
		if c.mutual {
			fmt.Fprintf(c.out, "created friendship between %v and %v OK\n", id2, id1)
		}
		fmt.Fprintf(c.out, "Process took %v \n", time.Since(start))
	}

	return nil
}
