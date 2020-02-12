package orbiter

import (
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/logging/format"
)

type DestroyFunc func() error

func Destroy(logger logging.Logger, gitClient *git.Client, adapt AdaptFunc) error {

	treeDesired, err := parse(gitClient)
	if err != nil {
		return err
	}

	treeCurrent := &Tree{}

	_, destroy, _, _, err := adapt(logger, treeDesired, treeCurrent)
	if err != nil {
		return err
	}

	if err := destroy(); err != nil {
		return err
	}

	logger.AddSideEffect(func(event bool, fields map[string]string) {
		if err := gitClient.UpdateRemote(format.CommitRecord(fields), git.File{
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
	}).Info(true, "Orb destroyed")
	return nil
}
