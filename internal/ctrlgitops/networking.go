package ctrlgitops

import (
	"context"
	"time"

	orbcfg "github.com/caos/orbos/pkg/orb"

	"github.com/caos/orbos/internal/operator/networking"
	"github.com/caos/orbos/internal/operator/networking/kinds/orb"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
)

func Networking(ctx context.Context, monitor mntr.Monitor, orbConfigPath string, k8sClient *kubernetes2.Client, binaryVersion *string) error {
	takeoffChan := make(chan struct{})
	go func() {
		takeoffChan <- struct{}{}
	}()

	gitClient := git.New(context.Background(), monitor, "orbos", "orbos@caos.ch")

	for range takeoffChan {
		orbConfig, err := orbcfg.ParseOrbConfig(orbConfigPath)
		if err != nil {
			monitor.Error(err)
			return err
		}

		if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
			monitor.Error(err)
			return err
		}

		takeoff := networking.Takeoff(monitor, gitClient, orb.AdaptFunc(ctx, binaryVersion, true), k8sClient)

		go func() {
			defer func() { monitor.RecoverPanic(recover()) }()
			started := time.Now()
			takeoff()

			monitor.WithFields(map[string]interface{}{
				"took": time.Since(started),
			}).Info("Iteration done")

			time.Sleep(time.Second * 30)
			takeoffChan <- struct{}{}
		}()
	}

	return nil
}
