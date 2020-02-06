package orbiter

import (
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
)

type DestroyFunc func() error

func Destroy(gitClient *git.Client, adapt AdaptFunc) error {

	treeDesired, err := parse(gitClient)
	if err != nil {
		return err
	}

	treeCurrent := &Tree{}
	_, destroy, _, _, err := adapt(treeDesired, treeCurrent)
	if err != nil {
		return err
	}

	if err := destroy(); err != nil {
		return err
	}

	if err := gitClient.UpdateRemote(git.File{
		Path:    "caos-internal/orbiter/current.yml",
		Content: []byte(""),
	}, git.File{
		Path:    "caos-internal/orbiter/node-agents-current.yml",
		Content: []byte(""),
	}, git.File{
		Path:    "caos-internal/orbiter/node-agents-desired.yml",
		Content: []byte(""),
	}, git.File{
		Path:    "orbiter.yml",
		Content: common.MarshalYAML(treeDesired),
	}); err != nil {
		panic(err)
	}
	return nil
}
