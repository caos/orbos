package cmds

import (
	"context"
	"strings"

	"gopkg.in/yaml.v3"

	cli "github.com/caos/orbos/pkg/cli"
	cli2 "github.com/caos/orbos/pkg/kubernetes/cli"

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
	deploy bool,
	verbose bool,
	version string,
	gitCommit string,
	kubeconfig string,
	gitOps bool,
	operators []string,
) error {

	if gitOps {
		if err := orbcfg.IsComplete(orbConfig); err != nil {
			return err
		}

		if err := cli.InitRepo(orbConfig, gitClient); err != nil {
			return err
		}

		if gitClient.Exists(git.OrbiterFile) && deployOperator(operators, "orbiter") {
			orbiterConfig := &ctrlgitops.OrbiterConfig{
				Recur:         recur,
				Deploy:        deploy,
				Verbose:       verbose,
				Version:       version,
				OrbConfigPath: orbConfig.Path,
				GitCommit:     gitCommit,
			}
			if err := ctrlgitops.Orbiter(ctx, monitor, orbiterConfig, gitClient); err != nil {
				return err
			}
		}
	}

	if !deploy {
		monitor.Info("Skipping operator deployments")
		return nil
	}

	k8sClient, err := cli2.Client(
		monitor,
		orbConfig,
		gitClient,
		kubeconfig,
		gitOps,
		false,
	)
	if err != nil {
		return err
	}

	if err := kubernetes.EnsureCaosSystemNamespace(monitor, k8sClient); err != nil {
		monitor.Info("failed to apply common resources into k8s-cluster")
		return err
	}

	if gitOps {

		orbConfigBytes, err := yaml.Marshal(orbConfig)
		if err != nil {
			return err
		}

		if err := kubernetes.EnsureOrbconfigSecret(monitor, k8sClient, orbConfigBytes); err != nil {
			monitor.Info("failed to apply configuration resources into k8s-cluster")
			return err
		}
	}

	if deployOperator(operators, "boom") {
		if err := deployBoom(monitor, gitClient, k8sClient, version, gitOps); err != nil {
			return err
		}
	}
	if deployOperator(operators, "networking") {
		return deployNetworking(monitor, gitClient, k8sClient, version, gitOps)
	}
	return nil
}

func deployOperator(arguments []string, operator string) bool {
	if len(arguments) == 0 {
		return true
	}

	for idx := range arguments {
		if strings.ToLower(arguments[idx]) == strings.ToLower(operator) {
			return true
		}
	}
	return false
}
