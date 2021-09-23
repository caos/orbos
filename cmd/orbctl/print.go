// Inspired by https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/

package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func PrintCommand(getRv GetRootValues) *cobra.Command {

	return &cobra.Command{
		Use:     "print <path>",
		Short:   "Print the files contents to stdout",
		Args:    cobra.ExactArgs(1),
		Example: `orbctl file print orbiter.yml`,
		RunE: func(cmd *cobra.Command, args []string) error {

			file := args[0]

			rv, err := getRv("print", "", map[string]interface{}{"file": file})
			if err != nil {
				return err
			}
			defer rv.ErrFunc(err)

			if !rv.Gitops {
				return errors.New("print command is only supported with the --gitops flag")
			}

			if err := initRepo(rv.OrbConfig, rv.GitClient); err != nil {
				return err
			}

			fmt.Print(string(rv.GitClient.Read(file)))
			return nil
		},
	}
}
