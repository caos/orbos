package main

import (
	"errors"

	"github.com/caos/orbos/pkg/tree"

	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/pkg/git"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/cli"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/orb"
	"github.com/caos/orbos/pkg/secret"

	"github.com/caos/orbos/internal/api"
	boomapi "github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/operator/orbiter"
	orbiterorb "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
)

func ConfigCommand(getRv GetRootValues) *cobra.Command {

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

		rv, _ := getRv()
		defer func() {
			err = rv.ErrFunc(err)
		}()

		if !rv.Gitops {
			return errors.New("configure command is only supported with the --gitops flag")
		}

		monitor := rv.Monitor
		orbConfig := rv.OrbConfig
		gitClient := rv.GitClient
		ctx := rv.Ctx

		if err := orb.ReconfigureAndClone(ctx, monitor, orbConfig, newRepoURL, newMasterKey, gitClient); err != nil {
			return err
		}

		rewriteKey := orbConfig.Masterkey
		if newMasterKey != "" {
			rewriteKey = newMasterKey
		}

		k8sClient, fromOrbiter, err := cli.Client(monitor, orbConfig, gitClient, rv.Kubeconfig, rv.Gitops)
		if err != nil {
			// ignore
			err = nil
		}

		if fromOrbiter {

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

			monitor.WithField("operator", "ORBITER").Info("Repopulating secrets")
			if err := secret.Rewrite(
				rewriteKey,
				func() error { return api.PushOrbiterDesiredFunc(gitClient, desired)(monitor) },
			); err != nil {
				return err
			}
		}

		configureOperator(
			"BOOM",
			monitor,
			gitClient,
			rewriteKey,
			func(client *git.Client) (b bool, err error) { return fromOrbiter, nil },
		)

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

			toolset, _, _, _, err := boomapi.ParseToolset(tree)
			if err != nil {
				return err
			}

			tree.Parsed = toolset
			if err := secret.Rewrite(
				rewriteKey,
				func() error { return api.PushBoomDesiredFunc(gitClient, tree)(monitor) },
			); err != nil {
				return err
			}
		}

		if k8sClient == nil {
			monitor.Info("Writing new orbconfig skipped as no kubernetes cluster connection is available")
			return nil
		}

		monitor.Info("Ensuring orbconfig in kubernetes cluster")

		orbConfigBytes, err := yaml.Marshal(orbConfig)
		if err != nil {
			return err
		}

		if err := kubernetes.EnsureOrbconfigSecret(monitor, k8sClient, orbConfigBytes); err != nil {
			monitor.Error(errors.New("failed to apply configuration resources into k8s-cluster"))
			return err
		}

		monitor.Info("Applied configuration resources")

		return nil
	}
	return cmd
}

func configureORBOSOperators(
	monitor mntr.Monitor,
	gitClient *git.Client,
	orbConfig *orb.Orb,
	rewriteKey string,
) error {

	if err := configureOperator(
		"ORBITER",
		monitor,
		gitClient,
		rewriteKey,
		api.ExistsOrbiterYml,
		func() (*tree.Tree, interface{}, error) {
			_, _, configure, _, desired, _, _, err := orbiter.Adapt(gitClient, monitor, make(chan struct{}), orbiterorb.AdaptFunc(
				labels.NoopOperator("ORBOS"),
				orbConfig,
				gitCommit,
				true,
				false,
				gitClient,
			))
			if err != nil {
				return nil, nil, err
			}

			return desired, desired.Parsed, configure(*orbConfig)
		},
		api.PushOrbiterDesiredFunc,
	); err != nil {
		return err
	}

	if err := configureOperator(
		"BOOM",
		monitor,
		gitClient,
		rewriteKey,
		api.ExistsBoomYml,
		func() (*tree.Tree, interface{}, error) {
			desired, err := api.ReadBoomYml(gitClient)
			if err != nil {
				return nil, nil, err
			}
			toolset, _, _, _, err := boomapi.ParseToolset(desired)
			return desired, toolset, err
		},
		api.PushBoomDesiredFunc,
	); err != nil {
		return err
	}

	return configureOperator(
		"networking operator",
		monitor,
		gitClient,
		rewriteKey,
		api.ExistsNetworkingYml,
		nil, // TODO!!!!
		api.PushNetworkingDesiredFunc,
	)
}

func configureOperator(
	operator string,
	monitor mntr.Monitor,
	gitClient *git.Client,
	rewriteKey string,
	exists func(*git.Client) (bool, error),
	configure func() (desired *tree.Tree, parsed interface{}, err error),
	push func(*git.Client, *tree.Tree) func(monitor mntr.Monitor) error,
) (func() error, error) {

	doIt, err := exists(gitClient)
	if err != nil || !doIt {
		return err
	}

	monitor.WithField("operator", operator).Info("Reconfiguring")

	tree, parsed, err := configure()
	if err != nil {
		return err
	}

	tree.Parsed = parsed
	return secret.Rewrite(
		rewriteKey,
		func() error { return push(gitClient, tree)(monitor) },
	)
}
