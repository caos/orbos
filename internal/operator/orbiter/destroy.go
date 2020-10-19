package orbiter

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
)

type DestroyFunc func() error

func NoopDestroy() error {
	return nil
}

func DestroyFuncGoroutine(query func() error) error {
	retChan := make(chan error)
	go func() {
		retChan <- query()
	}()
	return <-retChan
}

func Destroy(monitor mntr.Monitor, gitClient *git.Client, adapt AdaptFunc, finishedChan chan struct{}) error {
	treeDesired, err := api.ReadOrbiterYml(gitClient)
	if err != nil {
		return err
	}

	treeCurrent := &tree.Tree{}

	adaptFunc := func() (QueryFunc, DestroyFunc, ConfigureFunc, bool, map[string]*secret.Secret, error) {
		return adapt(monitor, finishedChan, treeDesired, treeCurrent)
	}

	_, destroy, _, _, _, err := AdaptFuncGoroutine(adaptFunc)
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
