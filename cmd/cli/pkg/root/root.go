package root

import (
	"context"
	"flag"

	"github.com/Lambels/relationer/internal/service"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type Config struct {
	Client service.GraphStore

	Verbose bool
	Path    string
}

func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.Verbose, "v", false, "log verbose output")
	fs.StringVar(&c.Path, "p", "http://localhost:8080", "api enpoint for relationer")
}

func (c *Config) Exec(context.Context, []string) error {
	// The root command has no meaning, so if it gets executed,
	// display the usage text to the user instead.
	return flag.ErrHelp
}

func New() (*ffcli.Command, *Config) {
	var cfg Config

	fs := flag.NewFlagSet("relationer", flag.ExitOnError)
	cfg.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "relationer",
		ShortUsage: "relationer is a cli tool to interact with the relationer server",
		FlagSet:    fs,
		Exec:       cfg.Exec,
	}, &cfg
}
