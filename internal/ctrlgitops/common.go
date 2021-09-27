package ctrlgitops

import (
	"context"
	"time"

	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/git"
)

func gitClient(monitor mntr.Monitor, task string) *git.Client {
	return git.New(context.Background(), monitor.WithField("task", task), "Boom", "boom@caos.ch")
}

func checks(monitor mntr.Monitor, client *git.Client) {
	defer func() { monitor.RecoverPanic(recover()) }()
	t := time.NewTicker(13 * time.Hour)
	for range t.C {
		if err := client.Check(); err != nil {
			monitor.Error(err)
		}
	}
}
