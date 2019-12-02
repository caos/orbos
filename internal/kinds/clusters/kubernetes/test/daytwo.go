package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/core/operator/test"
)

var (
	controlPlaneScale = 3
	workersScale      = 3
	kubernetesVersion = "v1.16.0"
)

func dayTwo(t *testing.T) (func(), func() error) {

	dayTwoCluster := fmt.Sprintf(`kind: orbiter.caos.ch/Orbiter
version: v1
deps:
  prod:
    kind: orbiter.caos.ch/KubernetesCluster
    version: v1
    spec:
      versions:
        kubernetes: %s
        orbiter: dev
        toolsop: comingSoon
      controlplane:
        updatesDisabled: false 
        provider: static
        nodes: %d
        pool: controlplane
      workers:
        static:
          updatesDisabled: false
          provider: static
          nodes: %d
          pool: isolated
#        lowcost:
#          updatesDisabled: false
#          provider: googlebelgium
#          nodes: 1
#          pool: centos
#        swiss:
#          updatesDisabled: false
#          provider: googlezurich
#          nodes: 2
#          pool: highstorage
    deps:
#      googlebelgium:
#        kind: orbiter.caos.ch/GCEProvider
#        version: v1
#        spec:
#          project: caos-240809
#          region: europe-west1
#          zone: europe-west1-b
#          pools:
#            highmem:
#              osImage: projects/ubuntu-os-cloud/global/images/ubuntu-1804-bionic-v20190918
#              minCPUCores: 2
#              minMemoryGB: 1
#              storageGB: 16
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
              externalIP: 10.61.0.206
              internalIP: 10.61.0.206
            - id: second
              hostname: a.centos
              externalIP: 10.61.0.207
              internalIP: 10.61.0.207
            - id: third
              hostname: another.ubuntu
              externalIP: 10.61.0.208
              internalIP: 10.61.0.208
            isolated:
            - id: fourth
              hostname: again.an.ubuntu
              externalIP: 10.61.0.209
              internalIP: 10.61.0.209
            - id: fifth
              hostname: again.an.ubuntu
              externalIP: 10.61.0.210
              internalIP: 10.61.0.210
            - id: sixth
              hostname: again.an.ubuntu
              externalIP: 10.61.0.211
              internalIP: 10.61.0.211
        deps:
          dynamic:
            kind: orbiter.caos.ch/DynamicLoadBalancer
            version: v1
#            deps:
#              kubeapi:
#                kind: orbiter.caos.ch/DynamicKubeAPILoadBalancer
#                version: v1
#                spec:
#                  ip: "10.61.0.254"
#              ingress:
#                kind: orbiter.caos.ch/DynamicKubeIngressLoadBalancer
#                version: v1
#                spec:
#                  ip: "10.61.0.253"
            spec:
              controlplane:
              - ip: "10.61.0.254"
                transport:
                - name: kubelet
                  sourcePort: 10248
                  healthchecks:
                    protocol: http
                    path: /healthz
                    code: 200
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
                - name: httpsingress
                  sourcePort: 443
                  healthchecks:
                    protocol: https
                    path: /healthz
                    code: 200
#                  destinations:
#                  - port: 443
#                    pool: isolated
                - name: httpingress
                  sourcePort: 80
                  healthchecks:
                    protocol: http
                    path: /healthz
                    code: 200
#                  destinations:
#                   - port: 80
#                     pool: isolated
`, kubernetesVersion, controlPlaneScale, workersScale)

	stop := make(chan struct{})
	iterations, cleanup, logger, err := test.Run(stop, "day two", t, dayTwoCluster, nil)

	if err != nil {
		t.Error(err)
		return func() {}, cleanup
	}

	return func() { dayTwoTester(t, logger, stop, iterations) }, cleanup
}

func dayTwoTester(t *testing.T, logger logging.Logger, stop chan struct{}, iterations chan *test.IterationResult) {
	testDayTwo, timer := dayTwoIterationTester(logger, t)

loop:
	for {
		select {
		case <-timer.C:
			t.Error("Timed out")
			break loop
		case it := <-iterations:
			if !testDayTwo(it) {
				break loop
			}
		}
	}
	stop <- struct{}{}
}

func dayTwoIterationTester(logger logging.Logger, t *testing.T) (func(it *test.IterationResult) bool, *time.Timer) {

	//	var errors int

	return func(it *test.IterationResult) bool {

		if it.Error != nil {
			t.Error(it.Error)
			return false
			//			if errors >= 2 {
			//				t.Error(it.Error)
			//				return false
			//			}
			//			errors++
			//			return true
		}

		//		errors = 0

		var allClusters int
		var uptodateClusters int
		for clusterName, cluster := range it.Current.Deps {

			if cluster.Current.Kind == "orbiter.caos.ch/KubernetesCluster" {
				allClusters++

				clusterLogger := logger.WithFields(map[string]interface{}{
					"cluster": clusterName,
				})
				computesAreUptodate := true
				var controlplane int
				var workers int
				for computeID, compute := range cluster.Current.State.Computes {
					state := compute.Software.Current.State
					nodeLogger := clusterLogger.WithFields(map[string]interface{}{
						"id":   computeID,
						"tier": compute.Metadata.Tier,
					})
					isRunning := compute.Status == "running"
					hasReadyNodeagent := state.Ready
					hasUptodateKubernetesVersion := state.Software.Kubelet.Version == kubernetesVersion
					if isRunning && hasReadyNodeagent && hasUptodateKubernetesVersion {
						switch compute.Metadata.Tier {
						case "controlplane":
							controlplane++
						case "workers":
							workers++
						default:
							t.Errorf("Unknown tier %s", compute.Metadata.Tier)
						}
						nodeLogger.Info("Node is up-to-date")
					} else {
						computesAreUptodate = false
						nodeLogger.WithFields(map[string]interface{}{
							"isRunning":                    isRunning,
							"hasReadyNodeagent":            hasReadyNodeagent,
							"hasUptodateKubernetesVersion": hasUptodateKubernetesVersion,
						}).Info("Node is not up-to-date yet")
					}
				}

				workersAreScaled := workers == workersScale
				controlplaneIsScaled := controlplane == controlPlaneScale
				clusterIsRunning := cluster.Current.State.Status == "running"
				if computesAreUptodate && workersAreScaled && controlplaneIsScaled && clusterIsRunning {
					uptodateClusters++
					clusterLogger.Info("Cluster is up-to-date")
				} else {
					clusterLogger.WithFields(map[string]interface{}{
						"computesAreUptodate":  computesAreUptodate,
						"workersAreScaled":     workersAreScaled,
						"controlplaneIsScaled": controlplaneIsScaled,
						"clusterIsRunning":     clusterIsRunning,
					}).Info("Cluster is not up-to-date yet")
				}
			}
		}

		if allClusters == uptodateClusters {
			logger.Info(fmt.Sprintf("Day two operations on %d clusters succeeded", allClusters))
			return false
		}

		logger.Info(fmt.Sprintf("%d/%d clusters are up to date", uptodateClusters, allClusters))
		return true

	}, time.NewTimer(30 * time.Minute)
}
