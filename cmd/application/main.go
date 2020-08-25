package main

import (
	"context"
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/boom"
	"github.com/caos/orbos/mntr"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	orbconfig := "/Users/benz/.orb/stefan-orbos-gce"

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	ensureClient := git.New(context.Background(), monitor.WithField("task", "ensure"), "Boom", "boom@caos.ch")
	queryClient := git.New(context.Background(), monitor.WithField("task", "query"), "Boom", "boom@caos.ch")

	takeoff, _ := boom.Takeoff(
		monitor,
		"./artifacts",
		true,
		orbconfig,
		ensureClient,
		queryClient,
	)
	takeoff()
}
