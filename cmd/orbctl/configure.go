package main

import (
	"errors"
	"io/ioutil"

	"github.com/caos/orbos/pkg/orb"

	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"

	boomapi "github.com/caos/orbos/internal/operator/boom/api"

	"github.com/caos/orbos/internal/start"

	"github.com/caos/orbos/internal/operator/orbiter"

	orbiterorb "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"

	"github.com/caos/orbos/internal/api"
	"github.com/spf13/cobra"
)

func ConfigCommand(rv RootValues) *cobra.Command {

	var (
		kubeconfig   string
		newMasterKey string
		newRepoURL   string
		cmd          = &cobra.Command{
			Use:     "configure",
			Short:   "Configures and reconfigures an orb",
			Long:    "Configures and reconfigures an orb",
			Aliases: []string{"reconfigure", "config", "reconfig"},
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&kubeconfig, "kubeconfig", "", "Needed in boom-only scenarios")
	flags.StringVar(&newMasterKey, "masterkey", "", "Reencrypts all secrets")
	flags.StringVar(&newRepoURL, "repourl", "", "Configures the repository URL")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		ctx, monitor, orbConfig, gitClient, errFunc, err := rv()
		if err != nil {
			return err
		}
		defer func() {
			err = errFunc(err)
		}()

		if err := orb.ReconfigureAndClone(ctx, monitor, orbConfig, newRepoURL, newMasterKey, gitClient); err != nil {
			return err
		}

		allKubeconfigs := make([]string, 0)
		foundOrbiter, err := api.ExistsOrbiterYml(gitClient)
		if err != nil {
			return err
		}

		rewriteKey := orbConfig.Masterkey
		if newMasterKey != "" {
			rewriteKey = newMasterKey
		}

		if foundOrbiter {

			_, _, configure, _, desired, _, _, err := orbiter.Adapt(gitClient, monitor, make(chan struct{}), orbiterorb.AdaptFunc(
				labels.NoopOperator("ORBOS"),
				orbConfig,
				gitCommit,
				true,
				false,
				gitClient,
			))
			if err != nil {
				return err
			}

			if err := configure(*orbConfig); err != nil {
				return err
			}

			monitor.Info("Repopulating orbiter secrets")
			if err := secret.Rewrite(
				monitor,
				gitClient,
				rewriteKey,
				desired,
				api.PushOrbiterDesiredFunc); err != nil {
				return err
			}

			monitor.Info("Reading kubeconfigs from orbiter.yml")
			kubeconfigs, err := start.GetKubeconfigs(monitor, gitClient, orbConfig, version)
			if err == nil {
				allKubeconfigs = append(allKubeconfigs, kubeconfigs...)
			}

		} else {
			monitor.Info("No orbiter.yml existent, reading kubeconfig from path provided as parameter")
			if kubeconfig == "" {
				return errors.New("error to change config as no kubeconfig is provided")
			}
			value, err := ioutil.ReadFile(kubeconfig)
			if err != nil {
				return err
			}
			allKubeconfigs = append(allKubeconfigs, string(value))
		}

		foundBoom, err := api.ExistsBoomYml(gitClient)
		if err != nil {
			return err
		}
		if foundBoom {
			monitor.Info("Repopulating boom secrets")

			tree, err := api.ReadBoomYml(gitClient)
			if err != nil {
				return err
			}

			toolset, _, _, _, _, err := boomapi.ParseToolset(tree)
			if err != nil {
				return err
			}

			tree.Parsed = toolset
			if err := secret.Rewrite(
				monitor,
				gitClient,
				rewriteKey,
				tree,
				api.PushBoomDesiredFunc); err != nil {
				return err
			}
		}

		for _, kubeconfig := range allKubeconfigs {
			k8sClient := kubernetes.NewK8sClient(monitor, &kubeconfig)
			if k8sClient.Available() {
				monitor.Info("Ensuring orbconfig in kubernetes cluster")
				if err := kubernetes.EnsureConfigArtifacts(monitor, k8sClient, orbConfig); err != nil {
					monitor.Error(errors.New("failed to apply configuration resources into k8s-cluster"))
					return err
				}

				monitor.Info("Applied configuration resources")
			} else {
				monitor.Info("No connection to the k8s-cluster possible")
			}
		}

		return nil
	}
	return cmd
}
