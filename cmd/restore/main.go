package main

import (
	"context"
	"fmt"
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/orb"
	orbconfig "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/start"
	"github.com/caos/orbos/mntr"
	"io/ioutil"
	"os"
)

func main() {
	orbConfigPath := "/Users/benz/.orb/zitadel-ch"

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	prunedPath := helpers.PruneHome(orbConfigPath)
	orbConfig, err := orbconfig.ParseOrbConfig(prunedPath)
	if err != nil {
		orbConfig = &orbconfig.Orb{Path: prunedPath}
	}

	ctx := context.Background()
	gitClient := git.New(ctx, monitor, "orbos", "orbos@caos.ch")

	if err := orbConfig.IsConnectable(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	found, err := api.ExistsZitadelYml(gitClient)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if found {
		zitadel, err := api.ReadZitadelYml(gitClient)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		list, err := orb.BackupListFunc()(monitor, zitadel)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		kubeconfig := "/Users/benz/.kube/zitadel-ch"

		data, err := ioutil.ReadFile(kubeconfig)
		if err != nil {
			monitor.Error(err)
			return
		}
		dummyKubeconfig := string(data)

		k8sClient := kubernetes.NewK8sClient(monitor, &dummyKubeconfig)
		/*
			kubeconfigs, err := start.GetKubeconfigs(monitor, gitClient)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}*/
		//for _, kubeconfig := range kubeconfigs {

		err = start.ZitadelRestore(monitor, orbConfig.Path, k8sClient, list[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		//}
	}
}
