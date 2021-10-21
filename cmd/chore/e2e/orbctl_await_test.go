package e2e_test

import (
	"fmt"
	"os"
	"time"

	"github.com/caos/orbos/internal/helpers"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type awaitEnsuredORBOS = func(expectMasters, expectWorkers uint8, k8sVersion string, timeout time.Duration)

func awaitEnsuredOrbiterFunc(orbctlGitops orbctlGitopsCmd, kubectl kubectlCmd, domain string) awaitEnsuredORBOS {

	return func(expectMasters, expectWorkers uint8, k8sVersion string, timeout time.Duration) {

		expectReadyPods, countReadyPods := assertReadyPods(kubectl, expectMasters, expectWorkers)

		print := printOperatorLogs(kubectl)

		type comparable struct {
			readyPods                                                 readyPods
			mastersDone, workersDone, nodeAgentsDone, currentMachines uint8
			clusterStatus                                             string
			vipAvailable, httpBinAvailable                            bool
		}

		Eventually(func() comparable {

			defer print()

			currentOrbiter := currentOrbiter(orbctlGitops).Clusters["k8s"].Current
			var (
				mastersDone uint8
				workersDone uint8
			)
			for machineID, machine := range currentOrbiter.Machines.M {
				if machine.Joined &&
					!machine.Rebooting &&
					machine.FirewallIsReady &&
					machine.Ready &&
					!machine.Unknown &&
					!machine.Updating {
					switch machine.Metadata.Tier {
					case kubernetes.Controlplane:
						mastersDone++
					case kubernetes.Workers:
						workersDone++
					default:
						Fail(fmt.Sprintf(`expected machine group to eighter be "%s" or "%s", but is "%s"`, kubernetes.Controlplane, kubernetes.Workers, machine.Metadata.Group))
					}
					continue
					fmt.Printf("machine %s is ready: %+v\n", machineID, *machine)
				}
				fmt.Printf("machine %s is not ready yet: %+v\n", machineID, *machine)
			}

			var nodeAgentsDone uint8
			for naID, na := range currentNodeagents(orbctlGitops).Current.NA {
				sw := na.Software
				if na.NodeIsReady &&
					sw.Kubeadm.Version == k8sVersion &&
					sw.Kubelet.Version == k8sVersion &&
					sw.Kubectl.Version == k8sVersion {
					nodeAgentsDone++
					continue
				}
				fmt.Printf("node agent %s is not done yet: ready=%t kubeadm=%s kubelet=%s kubectl=%s\n", naID, na.NodeIsReady, sw.Kubeadm.Version, sw.Kubelet.Version, sw.Kubectl.Version)
			}

			return comparable{
				mastersDone:      mastersDone,
				workersDone:      workersDone,
				clusterStatus:    currentOrbiter.Status,
				nodeAgentsDone:   nodeAgentsDone,
				currentMachines:  uint8(len(currentOrbiter.Machines.M)),
				readyPods:        countReadyPods(),
				vipAvailable:     checkVIPAvailability(orbctlGitops),
				httpBinAvailable: checkHTTPBinAvailability(domain),
			}
		}, timeout, 5).Should(Equal(comparable{
			mastersDone:      expectMasters,
			workersDone:      expectWorkers,
			clusterStatus:    "running",
			nodeAgentsDone:   expectMasters + expectWorkers,
			currentMachines:  expectMasters + expectWorkers,
			readyPods:        expectReadyPods,
			vipAvailable:     true,
			httpBinAvailable: true,
		}))
	}
}

type awaitUpdatedOrbiter func(patchPath, patchValue, expectK8sVersion string, expectMasters, expectWorkers uint8, timeout time.Duration)

func awaitUpdatedOrbiterFunc(orbctlGitops orbctlGitopsCmd, ExpectEnsuredOrbiter awaitEnsuredORBOS) awaitUpdatedOrbiter {
	return func(patchPath, patchValue, expectK8sVersion string, expectMasters, expectWorkers uint8, timeout time.Duration) {

		By(fmt.Sprintf("patching the orbiter.yml at %s using the value %s", patchPath, patchValue))

		session, err := gexec.Start(orbctlGitops("file", "patch", "orbiter.yml", patchPath, "--value", patchValue, "--exact"), GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 1*time.Minute).Should(gexec.Exit(0))

		By("waiting for ORBITER to ensure the result")

		ExpectEnsuredOrbiter(expectMasters, expectWorkers, expectK8sVersion, timeout)
	}
}

func printOperatorLogs(kubectl kubectlCmd) func() {
	from := time.Now()
	return func() {
		session, err := gexec.Start(kubectl("--namespace", "caos-system", "logs", "--selector", "app.kubernetes.io/name=orbiter", "--since-time", from.Format(time.RFC3339)), os.Stdout, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 30*time.Second).Should(FlakyExit())

		from = time.Now()
	}
}

func checkVIPAvailability(orbctl orbctlGitopsCmd) bool {
	provider := currentOrbiter(orbctl).Providers["providerundertest"]

	ep := provider.Current.Ingresses.Httpsingress

	msg, err := helpers.Check("https", ep.Location, ep.FrontendPort, "/ambassador/v0/check_ready", 200, false)
	fmt.Printf("ambassador ready check: %s: err: %v\n", msg, err)
	return err == nil
}

func checkHTTPBinAvailability(domain string) bool {
	msg, err := helpers.Check("https", fmt.Sprintf("httpbin.%s", domain), 443, "/get", 200, false)
	fmt.Printf("httpbin ready check: %s: err: %v\n", msg, err)
	return err == nil
}
