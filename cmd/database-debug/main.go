package main

import (
	"flag"
	"github.com/caos/orbos/pkg/kubernetes"
	"io/ioutil"

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

	kc, err := ioutil.ReadFile(helpers.PruneHome(*kubeconfig))
	if err != nil {
		panic(err)
	}

	if err := start.Database(
		monitor,
		helpers.PruneHome(*orbconfig),
		kubernetes.NewK8sClient(monitor, strPtr(string(kc))),
	); err != nil {
		panic(err)
	}
}

func strPtr(str string) *string {
	return &str
}
