package main

import (
	"context"
	"flag"
	"github.com/caos/orbos/internal/git"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/boom"
	"github.com/caos/orbos/mntr"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {

	orbconfig := flag.String("orbconfig", "~/.orb/config", "The orbconfig file to use")
	verbose := flag.Bool("verbose", false, "Print debug levelled logs")

	flag.Parse()

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	if *verbose {
		monitor = monitor.Verbose()
	}

	ensureClient := git.New(context.Background(), monitor.WithField("task", "ensure"), "Boom", "boom@caos.ch")
	queryClient := git.New(context.Background(), monitor.WithField("task", "query"), "Boom", "boom@caos.ch")

	takeoff, _ := boom.Takeoff(
		monitor,
		"./artifacts",
		true,
		helpers.PruneHome(*orbconfig),
		ensureClient,
		queryClient,
	)
	for {
		takeoff()
	}
}
