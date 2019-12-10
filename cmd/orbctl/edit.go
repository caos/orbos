// Inspired by https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/

package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/term"

	"github.com/caos/orbiter/internal/edge/git"
)

func editCommand(rv rootValues) *cobra.Command {
	return &cobra.Command{
		Use:   "edit [file]",
		Short: "Edit a file and push changes to the remote orb repository",
		Args:  cobra.ExactArgs(1),
		Example: `
orbctl --repourl git@github.com:example/my-orb.git \
       --repokey-file ~/.ssh/my-orb --masterkey 'my very secret key'
       edit desired.yml`,
		RunE: func(cmd *cobra.Command, args []string) error {

			_, logger, gitClient, _, errFunc := rv()
			if errFunc != nil {
				return errFunc(cmd)
			}

			if err := gitClient.Clone(); err != nil {
				panic(err)
			}

			file, err := gitClient.Read(args[0])
			if err != nil {
				panic(err)
			}

			edited, err := CaptureInputFromEditor(GetPreferredEditorFromEnvironment, bytes.NewReader(file))
			if err != nil {
				panic(err)
			}

			if bytes.Compare(file, edited) == 0 {
				logger.Info("No changes")
				return nil
			}

			_, err = gitClient.UpdateRemoteUntilItWorks(&git.File{
				Path: args[0],
				Overwrite: func(_ []byte) ([]byte, error) {
					return edited, nil
				},
				Force: true,
			})
			return err
		},
	}
}

// DefaultEditor is vim because we're adults ;)
const DefaultEditor = "vim"

// PreferredEditorResolver is a function that returns an editor that the user
// prefers to use, such as the configured `$EDITOR` environment variable.
type PreferredEditorResolver func() string

// GetPreferredEditorFromEnvironment returns the user's editor as defined by the
// `$EDITOR` environment variable, or the `DefaultEditor` if it is not set.
func GetPreferredEditorFromEnvironment() string {
	editor := os.Getenv("EDITOR")

	if editor == "" {
		return DefaultEditor
	}

	return editor
}

func resolveEditorArguments(executable string, filename string) []string {
	args := []string{filename}

	if strings.Contains(executable, "Visual Studio Code.app") {
		args = append([]string{"--wait"}, args...)
	}

	if strings.Contains(executable, "vim") {
		args = append([]string{"--not-a-term"}, args...)
	}

	// Other common editors

	return args
}

// OpenFileInEditor opens filename in a text editor.
func OpenFileInEditor(filename string, resolveEditor PreferredEditorResolver) error {
	// Get the full executable path for the editor.
	executable, err := exec.LookPath(resolveEditor())
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, resolveEditorArguments(executable, filename)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return (term.TTY{In: os.Stdin, TryDev: true}).Safe(cmd.Run)
}

// CaptureInputFromEditor opens a temporary file in a text editor and returns
// the written bytes on success or an error on failure. It handles deletion
// of the temporary file behind the scenes.
func CaptureInputFromEditor(resolveEditor PreferredEditorResolver, content io.Reader) ([]byte, error) {
	file, err := ioutil.TempFile(os.TempDir(), "*")
	if err != nil {
		return []byte{}, err
	}

	filename := file.Name()

	// Defer removal of the temporary file in case any of the next steps fail.
	defer os.Remove(filename)

	if _, err := io.Copy(file, content); err != nil {
		return []byte{}, err
	}

	if err = file.Close(); err != nil {
		return []byte{}, err
	}

	if err = OpenFileInEditor(filename, resolveEditor); err != nil {
		return []byte{}, err
	}

	return ioutil.ReadFile(filename)
}
