package orbiter

import (
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/mntr"
)

type DestroyFunc func() error

func Destroy(monitor mntr.Monitor, gitClient *git.Client, adapt AdaptFunc) error {

	treeDesired, err := parse(gitClient, "orbiter.yml")
	if err != nil {
		return err
	}

	treeCurrent := &Tree{}

	_, destroy, _, _, err := adapt(monitor, treeDesired[0], treeCurrent)
	if err != nil {
		return err
	}

	if err := destroy(); err != nil {
		return err
	}

	monitor.OnChange = func(evt string, fields map[string]string) {
		if err := gitClient.UpdateRemote(mntr.CommitRecord([]*mntr.Field{{Key: "evt", Value: evt}}), git.File{
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
	}
	monitor.Changed("Orb destroyed")
	return nil
}
