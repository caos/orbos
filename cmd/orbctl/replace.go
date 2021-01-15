package main

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

	"github.com/spf13/cobra"
)

func ReplaceCommand(getRv GetRootValues) *cobra.Command {
	return &cobra.Command{
		Use:   "replace",
		Short: "Replace a node with a new machine available in the same pool",
		Long:  "Pass machine ids as arguments, omit arguments for selecting machines interactively",
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			rv, err := getRv()
			if err != nil {
				return err
			}
			defer func() {
				err = rv.ErrFunc(err)
			}()

			orbConfig := rv.OrbConfig
			gitClient := rv.GitClient

			return requireMachines(monitor, gitClient, orbConfig, args, func(machine infra.Machine) (required bool, require func(), unrequire func()) {
				return machine.ReplacementRequired()
			})
		},
	}
}
