package main

import (
	"errors"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/secretfuncs"
	orbc "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/utils/orbgit"
	"github.com/spf13/cobra"
	"io/ioutil"
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
		ctx, monitor, orbConfig, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		gitClientConf := &orbgit.Config{
			Comitter:  "orbctl",
			Email:     "orbctl@caos.ch",
			OrbConfig: orbConfig,
			Action:    "config",
		}

		gitClient, cleanUp, err := orbgit.NewGitClient(ctx, monitor, gitClientConf)
		defer cleanUp()
		if err != nil {
			return err
		}

		allKubeconfigs := make([]string, 0)
		if existsFileInGit(gitClient, "orbiter.yml") {
			if masterkey != "" {
				if err := secret.Rewrite(
					monitor,
					gitClient,
					secretfuncs.GetRewrite(orbConfig, masterkey),
					"orbiter"); err != nil {
					panic(err)
				}
			}
		} else {
			if kubeconfig == "" {
				return errors.New("Error to change config as no kubeconfig is provided")
			}
			value, err := ioutil.ReadFile(kubeconfig)
			if err != nil {
				return err
			}
			allKubeconfigs = append(allKubeconfigs, string(value))
		}

		changedConfig := new(orbc.Orb)
		*changedConfig = *orbConfig
		if repoURL != "" {
			changedConfig.URL = repoURL
		}

		if masterkey != "" {
			changedConfig.Masterkey = masterkey
			if existsFileInGit(gitClient, "boom.yml") {
				if err := secret.Rewrite(
					monitor,
					gitClient,
					secretfuncs.GetRewrite(orbConfig, masterkey),
					"boom"); err != nil {
					return err
				}
			}

			if err := changedConfig.WriteBackOrbConfig(); err != nil {
				monitor.Info("failed to change local configuration")
				return err
			}
		}

		for _, kubeconfig := range allKubeconfigs {
			k8sClient := kubernetes.NewK8sClient(monitor, &kubeconfig)
			if k8sClient.Available() {
				if err := kubernetes.EnsureConfigArtifacts(monitor, k8sClient, changedConfig); err != nil {
					monitor.Info("failed to apply configuration resources into k8s-cluster")
					return err
				}

				monitor.Info("Applied configuration resources")
			}
		}

		return nil
	}
	return cmd
}
