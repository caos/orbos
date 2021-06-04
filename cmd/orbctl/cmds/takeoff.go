package cmds

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/pkg/kubernetes/cli"

	orbcfg "github.com/caos/orbos/pkg/orb"

	"github.com/caos/orbos/internal/ctrlgitops"

	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/kubernetes"
)

func Takeoff(
	monitor mntr.Monitor,
	ctx context.Context,
	orbConfig *orbcfg.Orb,
	gitClient *git.Client,
	recur bool,
	destroy bool,
	deploy bool,
	verbose bool,
	ingestionAddress string,
	version string,
	gitCommit string,
	kubeconfig string,
	gitOpsBoom bool,
	gitOpsNetworking bool,
) error {

	if !deploy {
		monitor.Info("Skipping operator deployments")
		return nil
	}

	getKubeClient := func() (*kubernetes.Client, bool, error) {
		return cli.Client(
			monitor,
			orbConfig,
			gitClient,
			kubeconfig,
			gitOpsBoom || gitOpsNetworking,
		)
	}

	k8sClient, fromOrbiter, err := getKubeClient()

	if !fromOrbiter && err != nil {
		return err
	}

	if fromOrbiter {
		err = nil
		orbiterConfig := &ctrlgitops.OrbiterConfig{
			Recur:            recur,
			Destroy:          destroy,
			Deploy:           deploy,
			Verbose:          verbose,
			Version:          version,
			OrbConfigPath:    orbConfig.Path,
			GitCommit:        gitCommit,
			IngestionAddress: ingestionAddress,
		}

		if err = ctrlgitops.Orbiter(ctx, monitor, orbiterConfig, gitClient); err != nil {
			return err
		}

		k8sClient, fromOrbiter, err = getKubeClient()
		if err != nil {
			return err
		}
	}

	if err := kubernetes.EnsureCaosSystemNamespace(monitor, k8sClient); err != nil {
		monitor.Info("failed to apply common resources into k8s-cluster")
		return err
	}

	if fromOrbiter || gitOpsBoom || gitOpsNetworking {

		orbConfigBytes, err := yaml.Marshal(orbConfig)
		if err != nil {
			return err
		}

		if err := kubernetes.EnsureOrbconfigSecret(monitor, k8sClient, orbConfigBytes); err != nil {
			monitor.Info("failed to apply configuration resources into k8s-cluster")
			return err
		}
	}

	if err := deployBoom(monitor, gitClient, k8sClient, version, gitOpsBoom); err != nil {
		return err
	}
	return deployNetworking(monitor, gitClient, k8sClient, version, gitOpsNetworking)
}
