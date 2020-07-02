package main

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/caos/orbos/internal/operator/orbiter"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"

	"github.com/caos/orbos/internal/ssh"
	"github.com/caos/orbos/internal/stores/github"

	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/secretfuncs"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/start"
	"github.com/spf13/cobra"
)

func ConfigCommand(rv RootValues) *cobra.Command {

	var (
		kubeconfig string
		masterkey  string
		repoURL    string
		cmd        = &cobra.Command{
			Use:   "config",
			Short: "Changes local and in-cluster of config",
			Long:  "Changes local and in-cluster of config",
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&kubeconfig, "kubeconfig", "", "Kubeconfig for in-cluster changes")
	flags.StringVar(&masterkey, "masterkey", "", "Masterkey to replace old masterkey in orbconfig")
	flags.StringVar(&repoURL, "repoURL", "", "Repository-URL to replace the old repository-URL in the orbconfig")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_, monitor, orbConfig, gitClient, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		var changes bool
		if masterkey != "" {
			monitor.Info("Change masterkey in current orbconfig")
			orbConfig.Masterkey = masterkey
			changes = true
		}
		if repoURL != "" {
			monitor.Info("Change repository url in current orbconfig")
			orbConfig.URL = repoURL
			changes = true
		}

		configureGit := func() error {
			return gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey))
		}

		// If the repokey already has read/write permissions, don't generate a new one.
		// This ensures git providers other than github keep being supported
		if err := configureGit(); err != nil {

			monitor.Info("Start connection with git-repository")

			dir := filepath.Dir(orbConfig.Path)

			deployKeyPrivLocal, deployKeyPub, err := ssh.Generate()
			if err != nil {
				panic(errors.New("failed to generate ssh key for deploy key"))
			}
			g := github.New(monitor).LoginOAuth(dir)
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

		if !changes {
			monitor.Info("No changes")
			return nil
		}

		monitor.Info("Writeback current orbconfig to local orbconfig")
		if err := orbConfig.WriteBackOrbConfig(); err != nil {
			monitor.Info("Failed to change local configuration")
			return err
		}

		allKubeconfigs := make([]string, 0)
		foundOrbiter, err := api.ExistsOrbiterYml(gitClient)
		if err != nil {
			return err
		}

		if foundOrbiter {

			_, _, configure, _, _, _, err := orbiter.Adapt(gitClient, monitor, make(chan struct{}), orb.AdaptFunc(
				orbConfig,
				gitCommit,
				true,
				false))
			if err != nil {
				return err
			}

			if err := configure(*orbConfig); err != nil {
				return err
			}

			monitor.Info("Reading kubeconfigs from orbiter.yml")
			kubeconfigs, err := start.GetKubeconfigs(monitor, gitClient)
			if err != nil {
				return err
			}
			allKubeconfigs = append(allKubeconfigs, kubeconfigs...)

			if masterkey != "" {
				monitor.Info("Read and rewrite orbiter.yml with new masterkey")
				if err := secret.Rewrite(
					monitor,
					gitClient,
					secretfuncs.GetRewrite(masterkey),
					"orbiter"); err != nil {
					panic(err)
				}
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

		if masterkey != "" {
			foundBoom, err := api.ExistsBoomYml(gitClient)
			if err != nil {
				return err
			}
			if foundBoom {
				monitor.Info("Read and rewrite boom.yml with new masterkey")
				if err := secret.Rewrite(
					monitor,
					gitClient,
					secretfuncs.GetRewrite(masterkey),
					"boom"); err != nil {
					return err
				}
			}
		}

		for _, kubeconfig := range allKubeconfigs {
			k8sClient := kubernetes.NewK8sClient(monitor, &kubeconfig)
			if k8sClient.Available() {
				monitor.Info("Ensure current orbconfig in kubernetes cluster")
				if err := kubernetes.EnsureConfigArtifacts(monitor, k8sClient, orbConfig); err != nil {
					monitor.Info("Failed to apply configuration resources into k8s-cluster")
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
