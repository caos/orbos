package main

import (
	"flag"

	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/orb"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/mntr"
)

func main() {
	var orbconfig, kubeconfig, metricsAddr string
	var verbose, gitopsmode bool

	flag.StringVar(&orbconfig, "orbconfig", "~/.orb/config", "The orbconfig file to use")
	flag.StringVar(&kubeconfig, "kc", "~/.kube/config", "The kubeconfig file to use")
	flag.BoolVar(&verbose, "verbose", false, "Print debug levelled logs")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&gitopsmode, "gitopsmode", false, "defines if the operator should run in gitops mode not crd mode")

	flag.Parse()

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	if verbose {
		monitor = monitor.Verbose()
	}

	if !gitopsmode {
		_, err := orb.ParseOrbConfig(helpers.PruneHome(orbconfig))
		if err != nil {
			panic(err)
		}

		if err := crdctrl.Start(monitor, "crdoperators", "./artifacts", metricsAddr, kubeconfig, crdctrl.Networking); err != nil {
			panic(err)
		}
	} else {
		cfg, err := clientgo.GetClusterConfig(monitor, kubeconfig)
		if err != nil {
			panic(err)
		}

		if err := gitopsctrl.Networking(
			monitor,
			helpers.PruneHome(orbconfig),
			kubernetes.NewK8sClientWithConfig(monitor, cfg),
			strPtr("networking-development"),
		); err != nil {
			panic(err)
		}
	}
}

func strPtr(str string) *string {
	return &str
}
