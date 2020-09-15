package main

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

	"github.com/spf13/cobra"
)

func RebootCommand(rv RootValues) *cobra.Command {
	return &cobra.Command{
		Use:   "reboot",
		Short: "Gracefully reboot machines",
		Long:  "Pass machine ids as arguments, omit arguments for selecting machines interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, monitor, orbConfig, gitClient, errFunc := rv()
			if errFunc != nil {
				return errFunc(cmd)
			}

			return requireMachines(monitor, gitClient, orbConfig, args, func(machine infra.Machine) (required bool, require func(), unrequire func()) {
				return machine.RebootRequired()
			})
		},
	}
}
