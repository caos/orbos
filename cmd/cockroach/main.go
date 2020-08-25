package main

import (
	"context"
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/orb"
	orbconfig "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/mntr"
	"io/ioutil"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
)

func main() {
	orbConfig := "/Users/benz/.orb/stefan-orbos-gce"

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	orbFile, err := orbconfig.ParseOrbConfig(orbConfig)
	if err != nil {
		monitor.Error(err)
	}

	ctx := context.Background()
	gitClient := git.New(ctx, monitor, "orbos", "orbos@caos.ch")

	if err := gitClient.Configure(orbFile.URL, []byte(orbFile.Repokey)); err != nil {
		os.Exit(1)
	}

	adapt := orb.AdaptFunc("", "networking", "zitadel", "database", "backup")
	kubeconfig := "/Users/benz/.kube/stefan-orbos-gce"

	data, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		monitor.Error(err)
		return
	}
	dummyKubeconfig := string(data)

	k8sClient := kubernetes.NewK8sClient(monitor, &dummyKubeconfig)
	//if err := k8sClient.RefreshLocal(); err != nil {
	//	return nil, nil, err
	//}

	takeoff := zitadel.Takeoff(
		monitor,
		gitClient,
		adapt,
		k8sClient,
	)

	if k8sClient.Available() {
		takeoff()

	}
}
