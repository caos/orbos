package main

import (
	"flag"

	"github.com/caos/orbos/internal/start"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/mntr"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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
		helpers.PruneHome(*kubeconfig),
		"zitadel",
	); err != nil {
		panic(err)
	}
}
