package main

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/boom/cmd"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/start"
	"github.com/caos/orbos/internal/utils/orbgit"
	"github.com/caos/orbos/internal/utils/random"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TakeoffCommand(rv RootValues) *cobra.Command {

	var (
		verbose          bool
		recur            bool
		destroy          bool
		deploy           bool
		kubeconfig       string
		ingestionAddress string
		cmd              = &cobra.Command{
			Use:   "takeoff",
			Short: "Launch orbos",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", false, "Ensure the desired state continously")
	flags.BoolVar(&deploy, "deploy", true, "Ensure Orbiter and Boom deployments continously")
	flags.StringVar(&ingestionAddress, "ingestion", "", "Ingestion API address")
	flags.StringVar(&kubeconfig, "kubeconfig", "", "Kubeconfig for boom deployment")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if recur && destroy {
			return errors.New("flags --recur and --destroy are mutually exclusive, please provide eighter one or none")
		}

		ctx, monitor, orbConfig, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		gitClientConf := &orbgit.Config{
			Comitter:  "orbctl",
			Email:     "orbctl@caos.ch",
			OrbConfig: orbConfig,
			Action:    "takeoff",
		}

		gitClient, cleanUp, err := orbgit.NewGitClient(ctx, monitor, gitClientConf, true)
		defer cleanUp()
		if err != nil {
			return err
		}

		allKubeconfigs := make([]string, 0)
		foundOrbiter, err := api.ExistsOrbiterYml(gitClient)
		if err != nil {
			return err
		}
		if foundOrbiter {
			orbiterConfig := &start.OrbiterConfig{
				Recur:            recur,
				Destroy:          destroy,
				Deploy:           deploy,
				Verbose:          verbose,
				Version:          version,
				OrbConfigPath:    orbConfig.Path,
				GitCommit:        gitCommit,
				IngestionAddress: ingestionAddress,
			}

			kubeconfigs, err := start.Orbiter(ctx, monitor, orbiterConfig, gitClient)
			if err != nil {
				return err
			}
			allKubeconfigs = append(allKubeconfigs, kubeconfigs...)
		} else {
			if kubeconfig == "" {
				return errors.New("Error to deploy BOOM as no kubeconfig is provided")
			}
			value, err := ioutil.ReadFile(kubeconfig)
			if err != nil {
				return err
			}
			allKubeconfigs = append(allKubeconfigs, string(value))
		}

		for _, kubeconfig := range allKubeconfigs {
			k8sClient := kubernetes.NewK8sClient(monitor, &kubeconfig)
			if k8sClient.Available() {
				if err := kubernetes.EnsureCommonArtifacts(monitor, k8sClient); err != nil {
					monitor.Info("failed to apply common resources into k8s-cluster")
					return err
				}
				monitor.Info("Applied common resources")

				if err := kubernetes.EnsureConfigArtifacts(monitor, k8sClient, orbConfig); err != nil {
					monitor.Info("failed to apply configuration resources into k8s-cluster")
					return err
				}
				monitor.Info("Applied configuration resources")
			} else {
				monitor.Info("Failed to connect to k8s")
			}

			if err := deployBoom(monitor, gitClient, &kubeconfig); err != nil {
				return err
			}
		}
		return nil
	}
	return cmd
}

func deployBoom(monitor mntr.Monitor, gitClient *git.Client, kubeconfig *string) error {
	foundBoom, err := api.ExistsBoomYml(gitClient)
	if err != nil {
		return err
	}
	if foundBoom {
		if err := cmd.Reconcile(monitor, kubeconfig, version); err != nil {
			return err
		}
	} else {
		monitor.Info("No BOOM deployed as no boom.yml present")
	}
	return nil
}

func StartOrbiter(rv RootValues) *cobra.Command {
	var (
		verbose          bool
		recur            bool
		destroy          bool
		deploy           bool
		ingestionAddress string
		cmd              = &cobra.Command{
			Use:   "orbiter",
			Short: "Launch an orbiter",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", true, "Ensure the desired state continously")
	flags.BoolVar(&deploy, "deploy", true, "Ensure Orbiter and Boom deployments continously")
	flags.StringVar(&ingestionAddress, "ingestion", "", "Ingestion API address")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if recur && destroy {
			return errors.New("flags --recur and --destroy are mutually exclusive, please provide eighter one or none")
		}

		ctx, monitor, orbConfig, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		gitClientConf := &orbgit.Config{
			Comitter:  "orbctl",
			Email:     "orbctl@caos.ch",
			OrbConfig: orbConfig,
			Action:    "takeoff",
		}

		gitClient, cleanUp, err := orbgit.NewGitClient(ctx, monitor, gitClientConf, true)
		defer cleanUp()
		if err != nil {
			return err
		}

		if err := gitClient.ReadCheck(); err != nil {
			monitor.Error(err)
			return err
		}

		if err := gitClient.WriteCheck(random.Generate()); err != nil {
			monitor.Error(err)
			return err
		}

		orbiterConfig := &start.OrbiterConfig{
			Recur:            recur,
			Destroy:          destroy,
			Deploy:           deploy,
			Verbose:          verbose,
			Version:          version,
			OrbConfigPath:    orbConfig.Path,
			GitCommit:        gitCommit,
			IngestionAddress: ingestionAddress,
		}

		_, err = start.Orbiter(ctx, monitor, orbiterConfig, gitClient)
		return err
	}
	return cmd
}

func StartBoom(rv RootValues) *cobra.Command {
	var (
		localmode bool
		cmd       = &cobra.Command{
			Use:   "boom",
			Short: "Launch a boom",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&localmode, "localmode", false, "Local mode for boom")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_, monitor, orbConfig, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		return start.Boom(monitor, orbConfig.Path, localmode, version)
	}
	return cmd
}
