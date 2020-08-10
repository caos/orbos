package main

import (
	"io/ioutil"
	"strings"
)

func run(branch, orbconfig string) error {
	newOrbctl, err := buildOrbctl(orbconfig)
	if err != nil {
		return err
	}

	kubeconfig, err := ioutil.TempFile("", "kubeconfig-*")
	if err != nil {
		return err
	}
	if err := kubeconfig.Close(); err != nil {
		return err
	}

	readKubeconfigTest, deleteKubeconfig := readKubeconfigTestFunc(kubeconfig.Name())
	defer deleteKubeconfig()

	branchParts := strings.Split(branch, "/")
	branch = branchParts[len(branchParts)-1:][0]

	if err := seq(newOrbctl, configureKubectl(kubeconfig.Name()),
		initORBITERTest,
		destroyTest,
		patchTestFunc("clusters.k8s.versions.orbiter", branch),
		bootstrapTest,
		readKubeconfigTest,
	); err != nil {
		return err
	}
	return nil
}

func seq(orbctl newOrbctlCommandFunc, kubectl newKubectlCommandFunc, fns ...func(newOrbctlCommandFunc, newKubectlCommandFunc) error) error {
	for _, fn := range fns {
		if err := fn(orbctl, kubectl); err != nil {
			return err
		}
	}
	return nil
}
