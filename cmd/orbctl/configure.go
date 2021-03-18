package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/pkg/kubernetes/cli"

	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/labels"
	secret2 "github.com/caos/orbos/pkg/secret"

	boomapi "github.com/caos/orbos/internal/operator/boom/api"

	"github.com/caos/orbos/internal/operator/orbiter"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"

	"github.com/caos/orbos/internal/ssh"
	"github.com/caos/orbos/internal/stores/github"

	"github.com/caos/orbos/internal/api"
	"github.com/spf13/cobra"
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

		if orbConfig.URL == "" && newRepoURL == "" {
			return errors.New("repository url is neighter passed by flag repourl nor written in orbconfig")
		}

		// TODO: Remove?
		if orbConfig.URL != "" && newRepoURL != "" && orbConfig.URL != newRepoURL {
			return fmt.Errorf("repository url %s is not reconfigurable", orbConfig.URL)
		}

		if orbConfig.Masterkey == "" && newMasterKey == "" {
			return errors.New("master key is neighter passed by flag masterkey nor written in orbconfig")
		}

		var changes bool
		if newMasterKey != "" {
			monitor.Info("Changing masterkey in current orbconfig")
			if orbConfig.Masterkey == "" {
				secret2.Masterkey = newMasterKey
			}
			orbConfig.Masterkey = newMasterKey
			changes = true
		}
		if newRepoURL != "" {
			monitor.Info("Changing repository url in current orbconfig")
			orbConfig.URL = newRepoURL
			changes = true
		}

		configureGit := func() error {
			return gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey))
		}

		// If the repokey already has read/write permissions, don't generate a new one.
		// This ensures git providers other than github keep being supported
		if err := configureGit(); err != nil {

			monitor.Info("Starting connection with git-repository")

			dir := filepath.Dir(orbConfig.Path)

			deployKeyPrivLocal, deployKeyPub, err := ssh.Generate()
			if err != nil {
				panic(errors.New("failed to generate ssh key for deploy key"))
			}
			g := github.New(monitor).LoginOAuth(ctx, dir)
			if g.GetStatus() != nil {
				return errors.New("failed github oauth login ")
			}
			repo, err := g.GetRepositorySSH(orbConfig.URL)
			if err != nil {
				return errors.New("failed to get github repository")
			}

			if err := g.EnsureNoDeployKey(repo).GetStatus(); err != nil {
				monitor.Error(errors.New("failed to clear deploy keys in repository"))
			}

			if err := g.CreateDeployKey(repo, deployKeyPub).GetStatus(); err != nil {
				return errors.New("failed to create deploy keys in repository")
			}
			orbConfig.Repokey = deployKeyPrivLocal

			if err := configureGit(); err != nil {
				return err
			}
			changes = true
		}

		if changes {
			monitor.Info("Writing local orbconfig")
			if err := orbConfig.WriteBackOrbConfig(); err != nil {
				monitor.Info("Failed to change local configuration")
				return err
			}
		}

		if err := gitClient.Clone(); err != nil {
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

			_, _, configure, _, desired, _, _, err := orbiter.Adapt(gitClient, monitor, make(chan struct{}), orb.AdaptFunc(
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
			if err := secret2.Rewrite(
				monitor,
				gitClient,
				rewriteKey,
				desired,
				api.PushOrbiterDesiredFunc); err != nil {
				return err
			}
			/*
				monitor.Info("Reading kubeconfigs from orbiter.yml")
				kubeconfigs, err := ctrlgitops.GetKubeconfigs(monitor, gitClient, orbConfig)
				if err == nil {
					allKubeconfigs = append(allKubeconfigs, kubeconfigs...)
				}
			*/
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

			toolset, _, _, _, _, _, err := boomapi.ParseToolset(tree)
			if err != nil {
				return err
			}

			tree.Parsed = toolset
			if err := secret2.Rewrite(
				monitor,
				gitClient,
				rewriteKey,
				tree,
				api.PushBoomDesiredFunc); err != nil {
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
