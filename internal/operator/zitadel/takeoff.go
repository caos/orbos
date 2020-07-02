package zitadel

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"io/ioutil"
)

func Takeoff(monitor mntr.Monitor, gitClient *git.Client, adapt AdaptFunc) func() {
	return func() {
		treeDesired, err := Parse(gitClient, "zitadel.yml")
		if err != nil {
			monitor.Error(err)
			return
		}
		treeCurrent := &tree.Tree{}

		data, err := ioutil.ReadFile("/Users/benz/.kube/config")
		dummyKubeconfig := string(data)
		k8sClient := kubernetes.NewK8sClient(monitor, &dummyKubeconfig)
		//if err := k8sClient.RefreshLocal(); err != nil {
		//	return nil, nil, err
		//}

		if !k8sClient.Available() {
			monitor.Error(err)
			return
		}

		query, _, err := adapt(monitor, treeDesired, treeCurrent)
		if err != nil {
			monitor.Error(err)
			return
		}

		ensure, err := query(k8sClient)
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
