package zitadel

import (
	"errors"
	"io/ioutil"

	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

func Takeoff(monitor mntr.Monitor, gitClient *git.Client, adapt AdaptFunc, kubeconfig string) func() {
	return func() {
		treeDesired, err := Parse(gitClient, "zitadel.yml")
		if err != nil {
			monitor.Error(err)
			return
		}
		treeCurrent := &tree.Tree{}

		var k8sClient *kubernetes.Client
		if kubeconfig != "" {
			data, err := ioutil.ReadFile(kubeconfig)
			if err != nil {
				monitor.Error(err)
				return
			}
			dummyKubeconfig := string(data)

			k8sClient = kubernetes.NewK8sClient(monitor, &dummyKubeconfig)
			//if err := k8sClient.RefreshLocal(); err != nil {
			//	return nil, nil, err
			//}
		} else {
			monitor.Error(errors.New("In cluster kubeconfig is not yet supported"))
			return
		}

		if !k8sClient.Available() {
			monitor.Error(err)
			return
		}

		query, _, err := adapt(monitor, treeDesired, treeCurrent)
		if err != nil {
			monitor.Error(err)
			return
		}

		ensure, err := query(k8sClient, map[string]interface{}{})
		if err != nil {
			monitor.Error(err)
			return
		}

		if err := ensure(k8sClient); err != nil {
			monitor.Error(err)
			return
		}
	}
}
