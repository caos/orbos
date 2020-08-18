package main

import (
	"fmt"

	"github.com/caos/orbos/internal/tree"

	"github.com/AlecAivazis/survey/v2"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

	"github.com/spf13/cobra"
)

func ExecCommand(rv RootValues) *cobra.Command {
	var (
		command string
		cmd     = &cobra.Command{
			Use:   "exec",
			Short: "Exec shell command on machine",
			Long:  "Exec shell command on machine",
			Args:  cobra.MaximumNArgs(1),
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&command, "command", "", "Command to be executed")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_, monitor, orbConfig, gitClient, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		return machines(monitor, gitClient, orbConfig, func(machineIDs []string, machines map[string]infra.Machine, _ *tree.Tree) error {

			machineID := ""
			if len(args) > 0 {
				machineID = args[0]
			} else {
				if err := survey.AskOne(&survey.Select{
					Message: "Select a machine:",
					Options: machineIDs,
				}, &machineID, survey.WithValidator(survey.Required)); err != nil {
					return err
				}
			}

			machine, found := machines[machineID]
			if !found {
				panic(fmt.Sprintf("Machine with ID %s unknown", machineID))
			}

			if command != "" {
				output, err := machine.Execute(nil, command)
				if err != nil {
					return err
				}
				fmt.Print(string(output))
			} else {
				if err := machine.Shell(); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return cmd
}
