package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

func runFunc(branch, orbconfig string, from int) func() error {
	return func() error {
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

		if err := seq(newOrbctl, configureKubectl(kubeconfig.Name()), from,
			/* 1 */ initORBITERTest,
			/* 2 */ destroyTest,
			/* 3 */ patchTestFunc("clusters.k8s.spec.versions.orbiter", branch),
			/* 4 */ bootstrapTest,
			/* 5 */ readKubeconfigTest,
			/* 6 */ waitTest(15*time.Second),
			/* 7 */ ensureORBITERTest(5*time.Minute),
		); err != nil {
			return err
		}
		return nil
	}
}

func seq(orbctl newOrbctlCommandFunc, kubectl newKubectlCommandFunc, from int, fns ...func(newOrbctlCommandFunc, newKubectlCommandFunc) error) error {
	var at int
	for _, fn := range fns {
		at++
		if at < from {
			fmt.Println("Skipping step", at)
			continue
		}
		if err := fn(orbctl, kubectl); err != nil {
			return err
		}
	}
	return nil
}
