// Inspired by https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/

package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func PrintCommand(rv RootValues) *cobra.Command {

	return &cobra.Command{
		Use:     "print <path>",
		Short:   "Print the files contents to stdout",
		Args:    cobra.ExactArgs(1),
		Example: `orbctl file print orbiter.yml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, orbConfig, gitClient, errFunc, err := rv()
			if err != nil {
				return err
			}
			defer func() {
				err = errFunc(err)
			}()

			if err := initRepo(orbConfig, gitClient); err != nil {
				return err
			}

			fmt.Print(string(gitClient.Read(args[0])))
			return nil
		},
	}
}
