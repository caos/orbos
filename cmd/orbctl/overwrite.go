// Inspired by https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/

package main

import (
	"fmt"

	"github.com/caos/orbos/internal/git"

	"github.com/spf13/cobra"
)

func OverwriteCommand(rv RootValues) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "overwrite <path>",
		Short:   "Create or replace the file",
		Args:    cobra.ExactArgs(1),
		Example: `orbctl file overwrite orbiter.yml`,
	}
	flags := cmd.Flags()

	var (
		value string
		file  string
		stdin bool
	)
	flags.StringVar(&value, "value", "", "Content value")
	flags.StringVarP(&file, "file", "s", "", "File containing the content value")
	flags.BoolVar(&stdin, "stdin", false, "The content value is read from standard input")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		content, err := key(value, file, stdin)
		if err != nil {
			return err
		}

		_, _, orbConfig, gitClient := rv()

		if err := initRepo(orbConfig, gitClient); err != nil {
			return err
		}

		return gitClient.UpdateRemote(fmt.Sprintf("Overwrite %s", args[0]), git.File{
			Path:    args[0],
			Content: []byte(content),
		})
	}

	return cmd
}
