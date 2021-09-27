package main

import (
	"errors"
	"fmt"

	"github.com/caos/orbos/v5/mntr"

	"github.com/caos/orbos/v5/pkg/tree"

	"github.com/AlecAivazis/survey/v2"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/clusters/core/infra"

	"github.com/spf13/cobra"
)

func ExecCommand(getRv GetRootValues) *cobra.Command {
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

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		machineID := ""
		if len(args) > 0 {
			machineID = args[0]
		}

		rv, err := getRv("exec", "", map[string]interface{}{"machine": machineID, "command": command != ""})
		if err != nil {
			return err
		}
		defer rv.ErrFunc(err)

		monitor := rv.Monitor
		orbConfig := rv.OrbConfig
		gitClient := rv.GitClient

		if !rv.Gitops {
			return mntr.ToUserError(errors.New("exec command is only supported with the --gitops flag and a committed orbiter.yml"))
		}

		return machines(monitor, gitClient, orbConfig, func(machineIDs []string, machines map[string]infra.Machine, _ *tree.Tree) error {

			if machineID == "" {
				if err := survey.AskOne(&survey.Select{
					Message: "Select a machine:",
					Options: machineIDs,
				}, &machineID, survey.WithValidator(survey.Required)); err != nil {
					return err
				}
			}

			machine, found := machines[machineID]
			if !found {
				return mntr.ToUserError(errors.New(fmt.Sprintf("Machine with ID %s unknown", machineID)))
			}

			if command != "" {
				output, err := machine.Execute(nil, command)
				if err != nil {
					return mntr.ToUserError(err)
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
