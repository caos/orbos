package main

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/spf13/cobra"
)

func machines(rv RootValues, cmd *cobra.Command, do func(machineIDs []string, machines map[string]infra.Machine) error) error {
	_, monitor, orbConfig, gitClient, errFunc := rv()
	if errFunc != nil {
		return errFunc(cmd)
	}

	if err := orbConfig.IsConnectable(); err != nil {
		return err
	}

	if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
		return err
	}

	if err := gitClient.Clone(); err != nil {
		return err
	}

	foundOrbiter, err := api.ExistsOrbiterYml(gitClient)
	if err != nil {
		return err
	}

	if foundOrbiter {
		monitor.Info("Reading machines from orbiter.yml")

		desired, err := api.ReadOrbiterYml(gitClient)
		if err != nil {
			return err
		}

		listMachines := orb.ListMachines()

		machineIDs, machines, err := listMachines(
			monitor,
			desired,
			orbConfig.URL,
		)

		return do(machineIDs, machines)
	}
	return nil
}
