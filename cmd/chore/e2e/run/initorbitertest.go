package main

import (
	"fmt"
	"os"
	"time"
)

func initORBITERTest(branch string) func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		print, err := orbctl()
		if err != nil {
			return err
		}

		print.Args = append(print.Args, "file", "print", "provider.yml")
		print.Stderr = os.Stderr

		var orbiterYml string
		if err := simpleRunCommand(print, time.NewTimer(30*time.Second), func(line string) (goon bool) {
			orbiterYml += fmt.Sprintf("    %s\n", line)
			return true
		}); err != nil {
			return err
		}

		orbiterYml = fmt.Sprintf(`kind: orbiter.caos.ch/Orb
version: v0
spec:
  verbose: false
clusters:
  k8s:
    kind: orbiter.caos.ch/KubernetesCluster
    version: v0
    spec:
      controlplane:
        updatesdisabled: false
        provider: provider-under-test
        nodes: 1
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
        kubernetes: v1.16.0
        orbiter: %s
      workers:
      - updatesdisabled: false
        provider: provider-under-test
        nodes: 1
        pool: application
      - updatesdisabled: false
        provider: provider-under-test
        nodes: 0
        pool: storage
providers:
  provider-under-test:
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
              code: 200`, branch, orbiterYml)

		fmt.Println(orbiterYml)

		overwrite, err := orbctl()
		if err != nil {
			return err
		}

		overwrite.Stderr = os.Stderr
		overwrite.Stderr = os.Stdout
		overwrite.Args = append(overwrite.Args, "file", "patch", "orbiter.yml", "--exact", "--value", orbiterYml)

		return overwrite.Run()
	}
}
