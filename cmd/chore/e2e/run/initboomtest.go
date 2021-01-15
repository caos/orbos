package main

import (
	"fmt"
	"os"
)

func initBOOMTest(branch string) func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		boomYml := fmt.Sprintf(`
apiVersion: boom.caos.ch/v1beta2
kind: Toolset
metadata:
  name: caos
  namespace: caos-system
spec:
  boomVersion: %s-dev
  postApply:
    deploy: false
  metricCollection:
    deploy: true
  logCollection:
    deploy: true
  nodeMetricsExporter:
    deploy: true
  systemdMetricsExporter:
    deploy: true
  monitoring:
    deploy: true
  apiGateway:
    deploy: true
    replicaCount: 3
  kubeMetricsExporter:
    deploy: true
  reconciling:
    deploy: true
  metricsPersisting:
    deploy: true
  logsPersisting:
    deploy: true`, branch)

		fmt.Println(boomYml)

		overwrite, err := orbctl()
		if err != nil {
			return err
		}

		overwrite.Stderr = os.Stderr
		overwrite.Stderr = os.Stdout
		overwrite.Args = append(overwrite.Args, "file", "patch", "boom.yml", "--exact", "--value", boomYml)

		return overwrite.Run()
	}
}
