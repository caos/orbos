package main

import (
	"flag"
	"github.com/pkg/errors"
	"os"

	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart/fetch"
	"github.com/caos/orbos/mntr"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	var basePath string

	verbose := flag.Bool("verbose", false, "Print logs for debugging")
	flag.StringVar(&basePath, "basepath", "./artifacts", "The local path where the base folder should be")
	flag.Parse()

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}
	if *verbose {
		monitor = monitor.Verbose()
	}

	// ctrl.SetLogger(monitor)

	if err := fetch.All(monitor, basePath); err != nil {
		monitor.Error(errors.Wrap(err, "unable to fetch charts"))
		os.Exit(1)
	}
}
