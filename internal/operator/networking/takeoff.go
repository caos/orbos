package networking

import (
	"errors"

	"github.com/caos/orbos/v5/internal/operator/core"
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/git"
	kubernetes2 "github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/tree"
)

func Takeoff(monitor mntr.Monitor, gitClient *git.Client, adapt core.AdaptFunc, k8sClient *kubernetes2.Client) func() {
	return func() {
		internalMonitor := monitor.WithField("operator", "networking")
		internalMonitor.Info("Takeoff")
		treeDesired, err := core.Parse(gitClient, "networking.yml")
		if err != nil {
			monitor.Error(err)
			return
		}
		treeCurrent := &tree.Tree{}

		if k8sClient == nil {
			internalMonitor.Error(errors.New("kubeclient is not available"))
			return
		}

		query, _, _, _, _, err := adapt(internalMonitor, treeDesired, treeCurrent)
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
