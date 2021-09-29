package cli

import (
	"bytes"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"io"
	"io/ioutil"
	"k8s.io/kubectl/pkg/util/term"
	"os"
	"os/exec"
	"strings"
)

func Edit(
	gitClient *git.Client,
	path string,
) error {

	edited, err := captureInputFromEditor(GetPreferredEditorFromEnvironment, bytes.NewReader(gitClient.Read(path)))
	if err != nil {
		panic(err)
	}

	return gitClient.UpdateRemote("File written by orbctl", func() []git.File {
		return []git.File{{
			Path:    path,
			Content: edited,
		}}
	})
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
		args = append([]string{"--not-a-term", "-c", "set nowrap"}, args...)
	}

	// Other common editors

	return args
}

// openFileInEditor opens filename in a text editor.
func openFileInEditor(filename string, resolveEditor PreferredEditorResolver) error {
	// Get the full executable path for the editor.
	executable, err := exec.LookPath(resolveEditor())
	if err != nil {
		return mntr.ToUserError(err)
	}

	cmd := exec.Command(executable, resolveEditorArguments(executable, filename)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return (term.TTY{In: os.Stdin, TryDev: true}).Safe(cmd.Run)
}

// captureInputFromEditor opens a temporary file in a text editor and returns
// the written bytes on success or an error on failure. It handles deletion
// of the temporary file behind the scenes.
func captureInputFromEditor(resolveEditor PreferredEditorResolver, content io.Reader) ([]byte, error) {
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

	if err = openFileInEditor(filename, resolveEditor); err != nil {
		return []byte{}, err
	}

	return ioutil.ReadFile(filename)
}
