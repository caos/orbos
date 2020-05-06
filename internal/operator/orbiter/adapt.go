package orbiter

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"gopkg.in/yaml.v3"
)

type AdaptFunc func(monitor mntr.Monitor, finishedChan chan bool, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, bool, error)

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
