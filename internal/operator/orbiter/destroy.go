package orbiter

import (
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
)

type DestroyFunc func() error

func Destroy(gitClient *git.Client, adapt AdaptFunc) error {

	treeDesired, treeSecrets, err := parse(gitClient)
	if err != nil {
		return err
	}

	treeCurrent := &Tree{}
	_, destroy, _, _, err := adapt(treeDesired, treeSecrets, treeCurrent)
	if err != nil {
		return err
	}

	if err := destroy(); err != nil {
		return err
	}

	if err := gitClient.UpdateRemote(git.File{
		Path:    "caos-internal/orbiter/current.yml",
		Content: common.MarshalYAML(treeCurrent),
	}, git.File{
		Path:    "secrets.yml",
		Content: common.MarshalYAML(treeSecrets),
	}); err != nil {
		panic(err)
	}
	return nil
}
