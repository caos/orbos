package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/caos/orbos/pkg/kubernetes/cli"

	orbcfg "github.com/caos/orbos/pkg/orb"

	"github.com/caos/orbos/cmd/orbctl/cmds"
	"github.com/caos/orbos/internal/ctrlcrd"
	"github.com/caos/orbos/internal/ctrlgitops"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TakeoffCommand(getRv GetRootValues) *cobra.Command {
	var (
		verbose          bool
		recur            bool
		destroy          bool
		deploy           bool
		ingestionAddress string
		gitOpsBoom       bool
		gitOpsNetworking bool
		cmd              = &cobra.Command{
			Use:   "takeoff",
			Short: "Launch an orbiter",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", false, "Ensure the desired state continously")
	flags.BoolVar(&deploy, "deploy", true, "Ensure Orbiter and Boom deployments continously")
	flags.BoolVar(&gitOpsBoom, "gitops-boom", false, "Ensure Boom runs in gitops mode")
	flags.BoolVar(&gitOpsNetworking, "gitops-networking", false, "Ensure Networking-operator runs in gitops mode")
	flags.StringVar(&ingestionAddress, "ingestion", "", "Ingestion API address")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		if recur && destroy {
			return errors.New("flags --recur and --destroy are mutually exclusive, please provide either one or none")
		}

		rv, err := getRv()
		if err != nil {
			return err
		}
		defer func() {
			err = rv.ErrFunc(err)
		}()

		orbConfig := rv.OrbConfig
		gitClient := rv.GitClient
		ctx := rv.Ctx

		return cmds.Takeoff(
			monitor,
			ctx,
			orbConfig,
			gitClient,
			recur,
			destroy,
			deploy,
			verbose,
			ingestionAddress,
			version,
			gitCommit,
			rv.Kubeconfig,
			rv.Gitops || gitOpsBoom,
			rv.Gitops || gitOpsNetworking,
		)
	}
	return cmd
}

func StartOrbiter(getRv GetRootValues) *cobra.Command {
	var (
		verbose          bool
		recur            bool
		destroy          bool
		deploy           bool
		ingestionAddress string
		pprof            bool
		cmd              = &cobra.Command{
			Use:   "orbiter",
			Short: "Launch an orbiter",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", true, "Ensure the desired state continously")
	flags.BoolVar(&deploy, "deploy", true, "Ensure Orbiter deployment continously")
	flags.BoolVar(&pprof, "pprof", false, "Start pprof to analyse memory usage")
	flags.StringVar(&ingestionAddress, "ingestion", "", "Ingestion API address")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		if recur && destroy {
			return errors.New("flags --recur and --destroy are mutually exclusive, please provide eighter one or none")
		}

		rv, err := getRv()
		if err != nil {
			return err
		}
		defer func() {
			err = rv.ErrFunc(err)
		}()

		monitor := rv.Monitor
		orbConfig := rv.OrbConfig
		gitClient := rv.GitClient
		ctx := rv.Ctx

		if err := orbcfg.IsComplete(orbConfig); err != nil {
			return err
		}

		if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
			return err
		}

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

		if pprof {
			go func() {
				log.Println(http.ListenAndServe("localhost:6060", nil))
			}()
		}

		return ctrlgitops.Orbiter(ctx, monitor, orbiterConfig, gitClient)
	}
	return cmd
}

func StartBoom(getRv GetRootValues) *cobra.Command {
	var (
		metricsAddr string
		cmd         = &cobra.Command{
			Use:   "boom",
			Short: "Launch a boom",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&metricsAddr, "metrics-addr", "", "The address the metric endpoint binds to.")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		rv, err := getRv()
		if err != nil {
			return err
		}
		defer func() {
			err = rv.ErrFunc(err)
		}()

		monitor := rv.Monitor
		orbConfig := rv.OrbConfig

		monitor.Info("Takeoff Boom")

		if rv.Gitops {
			return ctrlgitops.Boom(monitor, orbConfig.Path, version)
		} else {
			return ctrlcrd.Start(monitor, version, "/boom", metricsAddr, "", ctrlcrd.Boom)
		}
	}
	return cmd
}

func StartNetworking(getRv GetRootValues) *cobra.Command {
	var (
		metricsAddr string
		cmd         = &cobra.Command{
			Use:   "networking",
			Short: "Launch a networking operator",
			Long:  "Ensures a desired state of networking for an application",
		}
	)
	flags := cmd.Flags()
	flags.StringVar(&metricsAddr, "metrics-addr", "", "The address the metric endpoint binds to.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		rv, err := getRv()
		if err != nil {
			return err
		}
		defer func() {
			err = rv.ErrFunc(err)
		}()

		monitor := rv.Monitor
		orbConfig := rv.OrbConfig

		monitor.Info("Takeoff Networking")

		if rv.Gitops {

			k8sClient, err := cli.Client(monitor, orbConfig, rv.GitClient, rv.Kubeconfig, rv.Gitops, true)
			if err != nil {
				return err
			}

			return ctrlgitops.Networking(monitor, orbConfig.Path, k8sClient, &version)
		} else {
			return ctrlcrd.Start(monitor, version, "/boom", metricsAddr, rv.Kubeconfig, ctrlcrd.Networking)
		}
		return nil
	}
	return cmd
}
