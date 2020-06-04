package main

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/boom/cmd"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/start"
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
			Short: "Launch an orbiter",
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

		ctx, monitor, gitClient, orbFile, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		allKubeconfigs := make([]string, 0)
		if existsFileInGit(gitClient, "orbiter.yml") {
			kubeconfigs, err := start.Orbiter(ctx, monitor, recur, destroy, deploy, verbose, version, gitClient, orbFile, gitCommit, ingestionAddress)
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
				if err := kubernetes.EnsureCommonArtifacts(monitor, k8sClient, orbFile); err != nil {
					monitor.Info("failed to apply common resources into k8s-cluster")
					return err
				}
				monitor.Info("Applied common resources")
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
	if existsFileInGit(gitClient, "boom.yml") {
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

		ctx, monitor, gitClient, orbFile, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		_, err := start.Orbiter(ctx, monitor, recur, destroy, deploy, verbose, version, gitClient, orbFile, gitCommit, ingestionAddress)
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
		_, monitor, _, orbFile, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		return start.Boom(monitor, orbFile, localmode, version)
	}
	return cmd
}

func existsFileInGit(g *git.Client, path string) bool {
	if err := g.Clone(); err != nil {
		return false
	}

	of := g.Read(path)
	if of != nil && len(of) > 0 {
		return true
	}
	return false
}
