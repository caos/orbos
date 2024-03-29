package main

import (
	"errors"

	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

	"github.com/spf13/cobra"
)

func ReplaceCommand(getRv GetRootValues) *cobra.Command {
	return &cobra.Command{
		Use:   "replace",
		Short: "Replace a node with a new machine available in the same pool",
		Long:  "Pass machine ids as arguments, omit arguments for selecting machines interactively",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			node := ""
			if len(args) > 0 {
				node = args[0]
			}

			rv := getRv("replace", "", map[string]interface{}{"node": node})
			defer rv.ErrFunc(err)

			if !rv.Gitops {
				return mntr.ToUserError(errors.New("replace command is only supported with the --gitops flag and a committed orbiter.yml"))
			}

			return requireMachines(monitor, rv.GitClient, rv.OrbConfig, args, func(machine infra.Machine) (required bool, require func(), unrequire func()) {
				return machine.ReplacementRequired()
			})
		},
	}
}
