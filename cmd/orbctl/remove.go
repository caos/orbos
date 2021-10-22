// Inspired by https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/caos/orbos/pkg/git"
)

func RemoveCommand(getRv GetRootValues) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "remove <filepath>",
		Short:   "Remove file from git repository",
		Long:    "If the file doesn't exist, the command completes successfully",
		Args:    cobra.MinimumNArgs(1),
		Example: `orbctl file remove caos-internal/orbiter/current.yml`,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		filesStr := strings.Join(args, ",")

		rv, err := getRv("remove", "", map[string]interface{}{"files": filesStr})
		if err != nil {
			return err
		}
		defer rv.ErrFunc(err)

		if !rv.Gitops {
			return errors.New("remove command is only supported with the --gitops flag")
		}

		if err := initRepo(rv.OrbConfig, rv.GitClient); err != nil {
			return err
		}

		files := make([]git.File, len(args))
		for i := range args {
			files[i] = git.File{
				Path:    args[i],
				Content: nil,
			}
		}

		return rv.GitClient.UpdateRemote(fmt.Sprintf("Remove %s", filesStr), func() []git.File { return files })
	}

	return cmd
}