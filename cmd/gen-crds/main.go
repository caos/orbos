package main

import (
	"flag"
	"path/filepath"

	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
)

const (
	crdFolder = "crds"
)

func main() {
	var basePath string
	var boilerplatePath string
	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "If kubeconfig is provided, crds will get applied")
	flag.StringVar(&basePath, "basepath", "./artifacts", "The local path where the base folder should be")
	flag.StringVar(&boilerplatePath, "boilerplatepath", "./hack/boilerplate.go.txt", "The local path where the boilerplate text file lies")
	flag.Parse()

	if kubeconfig != "" {
		k8sClient, err := kubernetes.NewK8sClientPathBeforeInCluster(mntr.Monitor{}, kubeconfig)
		if err != nil {
			panic(err)
		}

		if k8sClient != nil {
			if err := kubernetes.ApplyCRDs(boilerplatePath, "./...", k8sClient); err != nil {
				panic(err)
			}
		}
	} else {
		if err := kubernetes.WriteCRDs(boilerplatePath, "./...", filepath.Join(basePath, crdFolder)); err != nil {
			panic(err)
		}
	}
}
