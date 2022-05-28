package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	addfriendship "github.com/Lambels/relationer/cmd/relationer/pkg/add_friendship"
	addperson "github.com/Lambels/relationer/cmd/relationer/pkg/add_person"
	getdepth "github.com/Lambels/relationer/cmd/relationer/pkg/get_depth"
	getfriendship "github.com/Lambels/relationer/cmd/relationer/pkg/get_friendship"
	getperson "github.com/Lambels/relationer/cmd/relationer/pkg/get_person"
	"github.com/Lambels/relationer/cmd/relationer/pkg/listen"
	"github.com/Lambels/relationer/cmd/relationer/pkg/root"
	"github.com/Lambels/relationer/internal/client"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	var (
		rootCmd, rootConf = root.New()
		createPerson      = addperson.New(rootConf, os.Stdout)
		createFriendship  = addfriendship.New(rootConf, os.Stdout)
		getDepth          = getdepth.New(rootConf, os.Stdout)
		getFriendship     = getfriendship.New(rootConf, os.Stdout)
		getPerson         = getperson.New(rootConf, os.Stdout)
		listen            = listen.New(rootConf, os.Stdout)
	)

	rootCmd.Subcommands = []*ffcli.Command{
		createPerson,
		createFriendship,
		getDepth,
		getFriendship,
		getPerson,
		listen,
	}

	if err := rootCmd.Parse(os.Args[1:]); err != nil {
		log.Fatalf("error during parse: %v\n", err)
	}

	client := client.NewClient(
		&http.Client{Timeout: 5 * time.Second},
		rootConf.Path,
	)

	rootConf.Client = client

	if err := rootCmd.Run(context.Background()); err != nil {
		log.Fatalf("%v\n", err)
	}
}
