package main

import (
	"fmt"

	"github.com/afiskon/promtail-client/promtail"
)

func initBOOMTest(logger promtail.Client, branch string) func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		boomYml := fmt.Sprintf(`
apiVersion: caos.ch/v1
kind: Boom
metadata:
  name: caos
  namespace: caos-system
spec:
  boomVersion: %s-dev
  postApply:
    deploy: false
  metricCollection:
    deploy: false
  logCollection:
    deploy: false
  nodeMetricsExporter:
    deploy: false
  systemdMetricsExporter:
    deploy: false
  monitoring:
    deploy: false
  apiGateway:
    deploy: true
    replicaCount: 1
  kubeMetricsExporter:
    deploy: false
  reconciling:
    deploy: false
  metricsPersisting:
    deploy: false
  logsPersisting:
    deploy: false`, branch)

		cmd, err := orbctl()
		if err != nil {
			return err
		}

		outWriter, outWrite := logWriter(logger.Infof)
		defer outWrite()
		cmd.Stdout = outWriter

		errWriter, errWrite := logWriter(logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

		cmd.Args = append(cmd.Args, "--gitops", "file", "patch", "boom.yml", "--exact", "--value", boomYml)

		return cmd.Run()
	}
}
