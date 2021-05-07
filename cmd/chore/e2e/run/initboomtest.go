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
