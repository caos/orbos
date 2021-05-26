package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"gopkg.in/yaml.v3"
)

var _ testFunc = writeInitialDesiredState

func writeInitialDesiredState(settings programSettings, expect *kubernetes.Spec) interactFunc {

	clusterSpec := fmt.Sprintf(`      controlplane:
        updatesdisabled: false
        provider: %s
        nodes: 3
        pool: management
        taints:
        - key: node-role.kubernetes.io/master
          effect: NoSchedule
      kubeconfig: null
      networking:
        dnsdomain: cluster.orbitertest
        network: calico
        servicecidr: 100.126.4.0/22
        podcidr: 100.127.224.0/20
      verbose: false
      versions:
        kubernetes: v1.19.9
        orbiter: %s
      workers:
      - updatesdisabled: false
        provider: %s
        nodes: 3
        pool: application`, settings.orbID, settings.artifactsVersion(), settings.orbID)

	if err := yaml.Unmarshal([]byte(clusterSpec), expect); err != nil {
		panic(err)
	}

	return func(_ uint8, orbctl newOrbctlCommandFunc) (time.Duration, checkCurrentFunc, error) {

		initCtx, initCancel := context.WithTimeout(settings.ctx, 30*time.Second)
		defer initCancel()

		var providerYml string
		if err := runCommand(settings, false, nil, func(line string) {
			providerYml += fmt.Sprintf("    %s\n", line)
		}, orbctl(initCtx), "--gitops", "file", "print", "provider.yml"); err != nil {
			return 0, nil, err
		}

		orbiterYml := fmt.Sprintf(`kind: orbiter.caos.ch/Orb
version: v0
spec:
  verbose: false
clusters:
  %s:
    kind: orbiter.caos.ch/KubernetesCluster
    version: v0
    spec:
%s
providers:
  %s:
%s
    loadbalancing:
      kind: orbiter.caos.ch/DynamicLoadBalancer
      version: v2
      spec:
        application:
        - ip: 192.168.122.11
          transport:
          - name: httpsingress
            frontendport: 443
            backendport: 30443
            backendpools:
            - application
            whitelist:
            - 0.0.0.0/0
            healthchecks:
              protocol: https
              path: /ambassador/v0/check_ready
              code: 200
            proxyprotocol: true
          - name: httpingress
            frontendport: 80
            backendport: 30080
            backendpools:
            - application
            whitelist:
            - 0.0.0.0/0
            healthchecks:
              protocol: http
              path: /ambassador/v0/check_ready
              code: 200
            proxyprotocol: true
        management:
        - ip: 192.168.122.10
          transport:
          - name: kubeapi
            frontendport: 6443
            backendport: 6666
            backendpools:
            - management
            whitelist:
            - 0.0.0.0/0
            healthchecks:
              protocol: https
              path: /healthz
              code: 200
`, settings.orbID, clusterSpec, settings.orbID, providerYml)
		orbiterCmd := orbctl(initCtx)
		orbiterCmd.Stdin = bytes.NewReader([]byte(orbiterYml))

		if err := runCommand(settings, true, nil, nil, orbiterCmd, "--gitops", "file", "patch", "orbiter.yml", "--exact", "--stdin"); err != nil {
			return 0, nil, err
		}

		return 0, nil, desireBOOMPlatform(settings, initCtx, orbctl, true)
	}
}

func desireBOOMPlatform(settings programSettings, ctx context.Context, orbctl newOrbctlCommandFunc, deploy bool) error {

	boomYml := fmt.Sprintf(`
apiVersion: caos.ch/v1
kind: Boom
metadata:
  name: caos
  namespace: caos-system
spec:
  boomVersion: %s
  postApply:
    deploy: false
  metricCollection:
    deploy: %t
  logCollection:
    deploy: %t
  nodeMetricsExporter:
    deploy: %t
  systemdMetricsExporter:
    deploy: %t
  monitoring:
    deploy: %t
  apiGateway:
    deploy: %t
    replicaCount: 1
  kubeMetricsExporter:
    deploy: %t
  reconciling:
    deploy: %t
  metricsPersisting:
    deploy: %t
  logsPersisting:
    deploy: %t
`,
		settings.artifactsVersion(),
		deploy,
		deploy,
		deploy,
		deploy,
		deploy,
		deploy,
		deploy,
		deploy,
		deploy,
		deploy,
	)
	boomCmd := orbctl(ctx)
	boomCmd.Stdin = bytes.NewReader([]byte(boomYml))

	return runCommand(settings, true, nil, nil, boomCmd, "--gitops", "file", "patch", "boom.yml", "--exact", "--stdin")

}
