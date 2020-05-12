package orbiter

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"gopkg.in/yaml.v3"
)

type AdaptFunc func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, bool, error)

func parse(gitClient *git.Client, files ...string) (trees []*tree.Tree, err error) {

	if err := gitClient.Clone(); err != nil {
		panic(err)
	}

	for _, file := range files {
		raw, err := gitClient.Read(file)
		if err != nil {
			return nil, err
		}

		tree := &tree.Tree{}
		if err := yaml.Unmarshal([]byte(raw), tree); err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}

	return trees, nil
}

type ret struct {
	query   QueryFunc
	destroy DestroyFunc
	migrate bool
	err     error
}

func AdaptFuncGoroutine(adapt func() (QueryFunc, DestroyFunc, bool, error)) (QueryFunc, DestroyFunc, bool, error) {
	retChan := make(chan ret)
	go func() {
		query, destroy, migrate, err := adapt()
		retChan <- ret{query, destroy, migrate, err}
	}()
	ret := <-retChan
	return ret.query, ret.destroy, ret.migrate, ret.err
}
