package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	addperson "github.com/Lambels/relationer/cmd/cli/pkg/add_person"
	"github.com/Lambels/relationer/cmd/cli/pkg/root"
	"github.com/Lambels/relationer/internal/client"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	var (
		rootCmd, rootConf = root.New()
		createPerson      = addperson.New(rootConf, os.Stdout)
	)

	rootCmd.Subcommands = []*ffcli.Command{
		createPerson,
	}

	if err := rootCmd.Parse(os.Args[1:]); err != nil {
		log.Fatalf("error during parse: %v\n", err)
	}

	client := client.NewClient(
		http.Client{Timeout: 5 * time.Second},
		rootConf.Path,
	)

	rootConf.Client = client

	if err := rootCmd.Run(context.Background()); err != nil {
		log.Fatalf("%v\n", err)
	}
}
