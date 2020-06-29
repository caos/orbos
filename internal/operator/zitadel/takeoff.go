package zitadel

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

func Takeoff(monitor mntr.Monitor, gitClient *git.Client) func() {
	return func() {
		treeDesired, err := Parse(gitClient, "zitadel.yml")
		if err != nil {
			monitor.Error(err)
			return
		}
		treeCurrent := &tree.Tree{}

		adapt := cockroachdb.AdaptFunc()

		query, _, err := adapt(monitor, treeDesired, treeCurrent)
		if err != nil {
			monitor.Error(err)
			return
		}

		ensure, err := query()
		if err != nil {
			monitor.Error(err)
			return
		}

		if err := ensure(); err != nil {
			monitor.Error(err)
			return
		}
	}
}
