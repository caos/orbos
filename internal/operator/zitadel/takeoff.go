package zitadel

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

func Takeoff(monitor mntr.Monitor, gitClient *git.Client, adapt AdaptFunc, k8sClient *kubernetes.Client) func() {
	return func() {
		internalMonitor := monitor.WithField("operator", "zitadel")
		internalMonitor.Info("Takeoff")
		treeDesired, err := Parse(gitClient, "zitadel.yml")
		if err != nil {
			monitor.Error(err)
			return
		}
		treeCurrent := &tree.Tree{}

		if !k8sClient.Available() {
			internalMonitor.Error(err)
			return
		}

		query, _, err := adapt(internalMonitor, treeDesired, treeCurrent)
		if err != nil {
			internalMonitor.Error(err)
			return
		}

		ensure, err := query(k8sClient, map[string]interface{}{})
		if err != nil {
			internalMonitor.Error(err)
			return
		}

		if err := ensure(k8sClient); err != nil {
			internalMonitor.Error(err)
			return
		}
	}
}
