package cli

import (
	"fmt"
	"github.com/caos/orbos/pkg/git"
	"strings"
)

func Remove(
	gitClient *git.Client,
	paths []string,
) error {
	files := make([]git.File, len(paths))
	for i := range paths {
		files[i] = git.File{
			Path:    paths[i],
			Content: nil,
		}
	}

	return gitClient.UpdateRemote(fmt.Sprintf("Remove %s", strings.Join(paths, ",")), func() []git.File { return files })
}
