package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/core/operator/test"
)

const dayOneDesired = `kind: orbiter.caos.ch/Orbiter
version: v1
deps:
  prod:
    kind: orbiter.caos.ch/KubernetesCluster
    version: v1
    spec:
      versions:
        kubernetes: "v1.15.0"
        orbiter: dev
        toolsop: comingSoon
      controlplane:
        updatesDisabled: false 
        provider: static
        nodes: 1
        pool: controlplane
      workers:
#        centos:
#          updatesDisabled: false
#          provider: googlebelgium
#          nodes: 1
#          pool: centos
#        swiss:
#          updatesDisabled: false
#          provider: googlezurich
#          nodes: 1
#          pool: highstorage
        static:
          updatesDisabled: false
          provider: static
          nodes: 1
          pool: isolated
    deps:
      static:
        kind: orbiter.caos.ch/StaticProvider
        version: v1
        spec:
          remoteuser: root
          remotePublicKeyPath: /root/.ssh/authorized_keys
          pools:
            controlplane:
            - id: first
              hostname: an.ubuntu
              externalIP: 10.61.0.176
              internalIP: 10.61.0.176
            - id: second
              hostname: a.centos
              externalIP: 10.61.0.177
              internalIP: 10.61.0.177
            - id: third
              hostname: another.ubuntu
              externalIP: 10.61.0.178
              internalIP: 10.61.0.178
            isolated:
            - id: fourth
              hostname: again.an.ubuntu
              externalIP: 10.61.0.179
              internalIP: 10.61.0.179
        deps:
          dynamic:
            kind: orbiter.caos.ch/DynamicLoadBalancer
            version: v1
            spec:
              controlplane:
              - ip: "10.61.0.254"
                transport:
                - name: kubeapi
                  sourcePort: 6443
                  healthchecks:
                    protocol: https
                    path: /healthz
                    code: 200
#                  destinations:
#                  - port: 6443
#                    pool: controlplane
              isolated:
              - ip: "10.61.0.253"
                transport:
                - name: dashboard
                  sourcePort: 443
                  destinations:
                  - port: 443
                    pool: controlplane
                    healthchecks:
                      protocol: https
                      path: /healthz
                      code: 200
              - ip: "10.61.0.252"
                transport:
                - name: httpsingress
                  sourcePort: 443
                  destinations:
                  - port: 443
                    pool: isolated
                    healthchecks:
                      protocol: https
                      path: /healthz
                      code: 200
                - name: httpingress
                  sourcePort: 80
                  destinations:
                  - port: 80
                    pool: controlplane
                    healthchecks:
                      protocol: https
                      path: /healthz
                      code: 200
                  - port: 8080
                    pool: isolated
                    healthchecks:
                      protocol: https
                      path: /healthz
                      code: 200
#          kubeapi:
#            kind: orbiter.caos.ch/ExternalLoadBalancer
#            version: v1
#            spec:
#              address: "10.61.0.254:6443"
#      googlebelgium:
#        kind: orbiter.caos.ch/GCEProvider
#        version: v1
#        spec:
#          project: caos-240809
#          region: europe-west1
#          zone: europe-west1-b
#          remoteuser: kube-operator
#          pools:
#            ubuntu:
#              osImage: projects/ubuntu-os-cloud/global/images/ubuntu-1804-bionic-v20190918
#              minCPUCores: 2
#              minMemoryGB: 1
#              storageGB: 16
#            centos:
#              osImage: projects/gce-uefi-images/global/images/centos-7-v20190916
#              minCPUCores: 2
#              minMemoryGB: 1
#              storageGB: 15
#      googlezurich:
#        kind: orbiter.caos.ch/GCEProvider
#        version: v1
#        spec:
#          project: caos-240809
#          region: europe-west6
#          zone: europe-west6-a
#          pools:
#            highstorage:
#              osImage: projects/ubuntu-os-cloud/global/images/ubuntu-1804-bionic-v20190918
#              minCPUCores: 2
#              minMemoryGB: 1
#              storageGB: 16 
`

func dayOne(t *testing.T) (func(), func() error) {

	stop := make(chan struct{})
	iterations, cleanup, logger, err := test.Run(stop, "dayOne", t, dayOneDesired, nil)

	if err != nil {
		t.Error(err)
		return func() {}, cleanup
	}

	return func() { dayOneTester(logger, t, stop, iterations) }, cleanup
}

func dayOneTester(logger logging.Logger, t *testing.T, stop chan struct{}, iterations chan *test.IterationResult) {

	dayOneIterationTester, timer := dayOneIterationTester(logger, t)

loop:
	for {
		select {
		case <-timer.C:
			t.Error("Timed out")
			break loop
		case it := <-iterations:
			if !dayOneIterationTester(it) {
				break loop
			}
		}
	}
	stop <- struct{}{}
}

func dayOneIterationTester(logger logging.Logger, t *testing.T) (func(it *test.IterationResult) bool, *time.Timer) {

	timer := time.NewTimer(5 * time.Minute)
	isFirst := true

	return func(it *test.IterationResult) bool {

		if it.Error != nil {
			t.Error(it.Error)
			return false
		}

		if isFirst {
			// Let node agents work
			timer.Reset(15 * time.Minute)
			isFirst = false
			return true
		}
		clusterKubeConfigs := make([]string, 0)

		var bootstrappedClusters int
		var allClusters int

	clusters:
		for clusterName, cluster := range it.Current.Deps {
			switch cluster.Current.Kind {
			case "orbiter.caos.ch/KubernetesCluster":
				allClusters++
				clusterKubeConfig := fmt.Sprintf("%s_kubeconfig", strings.ToLower(t.Name())+clusterName)
				// Success case first
				clusterIsRunning := cluster.Current.State.Status == "running"

				if !clusterIsRunning && cluster.Current.State.Status != "maintaining" {
					t.Errorf("Expected cluster %s to have status maintaining or running, but has %s", clusterName, cluster.Current.State.Status)
					return false
				}

				for computeName, compute := range cluster.Current.State.Computes {
					if compute.Status == "running" {
						clusterKubeConfigs = append(clusterKubeConfigs, clusterKubeConfig)
						continue
					}
					if compute.Status != "maintaining" {
						t.Errorf("Expected compute %s to have status maintaining, but has %s", computeName, compute.Status)
						return false
					}
				}

				for _, kubeconfig := range clusterKubeConfigs {
					if clusterKubeConfig == kubeconfig && clusterIsRunning {
						logger.Info(fmt.Sprintf("Cluster %s is bootstrapped", clusterName))
						bootstrappedClusters++
						continue clusters
					}
					logger.Info(fmt.Sprintf("Cluster %s is not running yet, instead it is %s", clusterName, cluster.Current.State.Status))
				}
			default:
				t.Errorf("unexpected cluster kind %s", cluster.Current.Kind)
				return false
			}
		}

		for _, kubeconfig := range clusterKubeConfigs {
			if _, ok := it.Secrets[kubeconfig]; !ok {
				t.Errorf("finding kubeconfig %s in secrets file failed", kubeconfig)
				return false
			}
		}

		if allClusters > 0 && allClusters == bootstrappedClusters {
			logger.Info(fmt.Sprintf("Bootstrapping of %d clusters succeeded", bootstrappedClusters))
			return false
		}

		logger.Info(fmt.Sprintf("%d/%d clusters are bootstrapped", bootstrappedClusters, allClusters))
		return true
	}, timer
}
