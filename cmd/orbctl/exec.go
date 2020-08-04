package main

import (
	"fmt"

	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func ExecCommand(rv RootValues) *cobra.Command {
	var (
		command   string
		machineID string
		cmd       = &cobra.Command{
			Use:   "exec",
			Short: "Exec shell command on machine",
			Long:  "Exec shell command on machine",
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&machineID, "machine", "", "ID of the machine to connect to")
	flags.StringVar(&command, "command", "", "Command to be executed")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_, monitor, orbConfig, gitClient := rv()

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
		}
		return nil
	}
	return cmd
}
