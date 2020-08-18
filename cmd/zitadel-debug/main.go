package main

import (
	"flag"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"

	"github.com/caos/orbos/internal/start"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/mntr"
)

func main() {

	orbconfig := flag.String("orbconfig", "~/.orb/config", "The orbconfig file to use")
	kubeconfig := flag.String("kubeconfig", "~/.kube/config", "The kubeconfig file to use")
	verbose := flag.Bool("verbose", false, "Print debug levelled logs")

	flag.Parse()

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	if *verbose {
		monitor = monitor.Verbose()
	}

	if err := start.Zitadel(
		monitor,
		helpers.PruneHome(*orbconfig),
		kubernetes.NewK8sClient(monitor, strPtr(helpers.PruneHome(*kubeconfig))),
	); err != nil {
		panic(err)
	}
}

func strPtr(str string) *string {
	return &str
}
