package main

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func RebootCommand(rv RootValues) *cobra.Command {
	var (
		command   string
		machineID string
		cmd       = &cobra.Command{
			Use:   "reboot",
			Short: "Gracefully reboot machines",
			Long:  "Gracefully reboot machines",
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&machineID, "machine", "", "ID of the machine to connect to")
	flags.StringVar(&command, "command", "", "Command to be executed")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return machines(rv, cmd, func(machineIDs []string, machines map[string]infra.Machine) error {

			if machineID == "" {
				prompt := promptui.Select{
					Label: "Select machine",
					Items: machineIDs,
				}

				_, result, err := prompt.Run()
				if err != nil {
					return err
				}
				machineID = result
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
