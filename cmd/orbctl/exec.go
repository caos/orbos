package main

import (
	"fmt"
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/internal/utils/orbgit"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func Exec(rv RootValues) *cobra.Command {
	var (
		interactive bool
		command     string
		machineID   string
		cmd         = &cobra.Command{
			Use:   "exec",
			Short: "Exec shell command on machine",
			Long:  "Exec shell command on machine",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&interactive, "interactive", false, "Use the created connection interactive")
	flags.StringVar(&machineID, "machine", "", "ID of the machine to connect to")
	flags.StringVar(&command, "command", "", "Command to be executed")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_, monitor, orbConfig, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		ctx, monitor, orbConfig, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		gitClientConf := &orbgit.Config{
			Comitter:  "orbctl",
			Email:     "orbctl@caos.ch",
			OrbConfig: orbConfig,
			Action:    "exec",
		}

		gitClient, cleanUp, err := orbgit.NewGitClient(ctx, monitor, gitClientConf, true)
		defer cleanUp()
		if err != nil {
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
				output, err := machine.Execute(nil, nil, command)
				if err != nil {
					return err
				}
				monitor.Info(string(output))
			} else {
				if err := machine.Shell(nil); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return cmd
}
