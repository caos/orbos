package main

import (
	"fmt"

	"github.com/caos/orbos/internal/api"

	"github.com/caos/orbos/internal/tree"

	"github.com/AlecAivazis/survey/v2"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

	"github.com/spf13/cobra"
)

func RebootCommand(rv RootValues) *cobra.Command {
	var (
		command string
		cmd     = &cobra.Command{
			Use:   "reboot",
			Short: "Gracefully reboot machines",
			Long:  "Gracefully reboot machines",
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&command, "command", "", "Command to be executed")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_, monitor, orbConfig, gitClient, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		return machines(monitor, gitClient, orbConfig, func(machineIDs []string, machines map[string]infra.Machine, desired *tree.Tree) error {

			if len(args) <= 0 {
				if err := survey.AskOne(&survey.MultiSelect{
					Message: "Select machines:",
					Options: machineIDs,
				}, &args, survey.WithValidator(survey.Required)); err != nil {
					return err
				}
			}

			var push bool
			for _, arg := range args {
				machine, found := machines[arg]
				if !found {
					panic(fmt.Sprintf("Machine with ID %s unknown", arg))
				}
				required, require, _ := machine.RebootRequired()
				if !required {
					require()
					push = true
				}
			}

			if !push {
				monitor.Info("Nothing changed")
				return nil
			}
			return api.PushOrbiterYml(monitor, "Update orbiter.yml", gitClient, desired)
		})
	}
	return cmd
}
