package orbiter

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"gopkg.in/yaml.v3"
)

type AdaptFunc func(monitor mntr.Monitor, finishedChan chan bool, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, bool, error)

func Parse(gitClient *git.Client, files ...string) (trees []*tree.Tree, err error) {

	if err := gitClient.Clone(); err != nil {
		return nil, err
	}

	for _, file := range files {
		tree := &tree.Tree{}
		if err := yaml.Unmarshal(gitClient.Read(file), tree); err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}

	return trees, nil
}

type retAdapt struct {
	query   QueryFunc
	destroy DestroyFunc
	migrate bool
	err     error
}

func AdaptFuncGoroutine(adapt func() (QueryFunc, DestroyFunc, bool, error)) (QueryFunc, DestroyFunc, bool, error) {
	retChan := make(chan retAdapt)
	go func() {
		query, destroy, migrate, err := adapt()
		retChan <- retAdapt{query, destroy, migrate, err}
	}()
	ret := <-retChan
	return ret.query, ret.destroy, ret.migrate, ret.err
}
