package main

import (
	"github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/start"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

func TakeoffCommand(rv RootValues) *cobra.Command {
	var (
		verbose          bool
		recur            bool
		destroy          bool
		deploy           bool
		ingestionAddress string
		cmd              = &cobra.Command{
			Use:   "takeoff",
			Short: "Take off with orbiter and boom",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", false, "Ensure the desired state continously")
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

		if err := start.Orbiter(ctx, monitor, recur, destroy, deploy, verbose, version, gitClient, orbFile, gitCommit, ingestionAddress); err != nil {
			return err
		}

		path := "orbiter.k8s.kubeconfig"
		secretFunc := func(operator string) secret.Func {
			if operator == "boom" {
				return api.SecretFunc(orbFile)
			} else if operator == "orbiter" {
				return orb.SecretsFunc(orbFile)
			}
			return nil
		}

		value, err := secret.Read(
			monitor,
			gitClient,
			secretFunc,
			path)
		if err != nil {
			monitor.Info("Failed to get kubeconfig")
			os.Exit(1)
		}
		monitor.Info("Read kubeconfig for boom deployment")

		k8sClient := kubernetes.NewK8sClient(monitor, &value)

		if k8sClient.Available() {
			if err := kubernetes.EnsureBoomArtifacts(monitor, k8sClient, "v0.10.7"); err != nil {
				monitor.Info("failed to deploy boom into k8s-cluster")
				return err
			}
			monitor.Info("Deployed boom")
		} else {
			monitor.Info("Failed to connect to k8s")
		}

		return nil
	}
	return cmd
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

		return start.Orbiter(ctx, monitor, recur, destroy, deploy, verbose, version, gitClient, orbFile, gitCommit, ingestionAddress)
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
		ctx, monitor, _, orbFile, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		return start.Boom(ctx, monitor, orbFile, localmode)
	}
	return cmd
}
