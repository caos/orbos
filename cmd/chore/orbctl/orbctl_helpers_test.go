package orbctl_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega/types"

	"github.com/caos/orbos/internal/operator/common"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/cmd/chore/orbctl"
	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/pkg/orb"
)

func parseUint8(t *testing.T, val string) uint8 {
	parsed, err := strconv.ParseInt(val, 10, 8)
	if err != nil {
		t.Fatal(err)
	}
	return uint8(parsed)
}

func boolEnv(val string) bool {
	if val == "" {
		return false
	}
	value, err := strconv.ParseBool(val)
	Expect(err).ToNot(HaveOccurred())
	return value
}

func calcOrbID(orbconfig string) string {
	orbCfg, err := orb.ParseOrbConfig(helpers.PruneHome(orbconfig))
	Expect(err).ToNot(HaveOccurred())

	Expect(orb.IsComplete(orbCfg)).ToNot(HaveOccurred())

	return strings.ToLower(strings.Split(strings.Split(orbCfg.URL, "/")[1], ".")[0])
}

type orbctlGitopsCmd func(args ...string) *exec.Cmd

func orbctlGitopsFunc(orbconfig string) orbctlGitopsCmd {
	cmdFunc, error := orbctl.Command(false, true, false, "")
	Expect(error).ToNot(HaveOccurred())
	return func(args ...string) *exec.Cmd {
		cmd := cmdFunc(context.Background())
		cmd.Args = append(cmd.Args, append([]string{"--disable-analytics", "--gitops", "--orbconfig", orbconfig}, args...)...)
		return cmd
	}
}

func memoizeUnmarshalE2eYml(orbctl orbctlGitopsCmd) func(interface{}) {

	var bytes []byte
	return func(into interface{}) {

		if bytes != nil {
			unmarshalYaml(bytes, into)
			return
		}

		By("fetching the file e2e.yml from git")

		session, err := gexec.Start(orbctl("file", "print", "e2e.yml"), GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5).Should(gexec.Exit(0))

		bytes = session.Out.Contents()
		unmarshalYaml(bytes, into)
	}
}

type kubectlCmd func(...string) *exec.Cmd

func memoizeKubecltCmd(kubectlPath string, orbctl orbctlGitopsCmd) kubectlCmd {
	var read bool

	return func(args ...string) *exec.Cmd {
		cmd := exec.Command("kubectl", append([]string{"--kubeconfig", kubectlPath}, args...)...)

		if read {
			return cmd
		}

		file, err := os.Create(kubectlPath)
		Expect(err).ToNot(HaveOccurred())
		defer file.Close()

		session, err := gexec.Start(orbctl("readsecret", "orbiter.k8s.kubeconfig.encrypted"), file, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5).Should(gexec.Exit(0))

		read = true

		return cmd
	}
}

func unmarshalYaml(content []byte, into interface{}) {
	Expect(yaml.Unmarshal(content, into)).To(Succeed())
}

func unmarshalStdoutYaml(cmd *exec.Cmd, into interface{}, flaky bool) {

	var matcher types.GomegaMatcher = gexec.Exit(0)
	if flaky {
		matcher = FlakyExit(0)
	}

	session, err := gexec.Start(cmd, nil, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 2*time.Minute).Should(matcher)

	unmarshalYaml(session.Out.Contents(), into)
}

func readyPods(kubectl kubectlCmd, namespace, selector string) (readyPodsCount uint8) {

	pods := struct {
		Items []struct {
			Metadata struct {
				Name string
			}
			Status struct {
				Conditions []struct {
					Type   string
					Status string
				}
			}
		}
	}{}

	args := []string{
		"get", "pods",
		"--namespace", namespace,
		"--output", "yaml",
	}

	if selector != "" {
		args = append(args, "--selector", selector)
	}

	unmarshalStdoutYaml(kubectl(args...), &pods, true)

	for i := range pods.Items {
		pod := pods.Items[i]
		for j := range pod.Status.Conditions {
			condition := pod.Status.Conditions[j]
			if condition.Type != "Ready" {
				continue
			}
			if condition.Status == "True" {
				readyPodsCount++
				break
			}
		}
	}

	return readyPodsCount
}

func printOperatorLogs(kubectl kubectlCmd) func() {
	from := time.Now()
	return func() {
		session, err := gexec.Start(kubectl("--namespace", "caos-system", "logs", "--selector", "app.kubernetes.io/name=orbiter", "--since-time", from.Format(time.RFC3339)), os.Stdout, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())

		from = time.Now()
	}
}

func currentOrbiter(orbctlGitops orbctlGitopsCmd) (currentOrbiter struct {
	Clusters map[string]struct {
		Current kubernetes.CurrentCluster
	}
	Providers map[string]struct {
		Current struct {
			Ingresses struct {
				Httpsingress infra.Address
				Httpingress  infra.Address
				Kubeapi      infra.Address
			}
		}
	}
}) {

	unmarshalStdoutYaml(orbctlGitops("file", "print", "caos-internal/orbiter/current.yml"), &currentOrbiter, false)
	return currentOrbiter
}

func currentNodeagents(orbctlGitops orbctlGitopsCmd) *common.NodeAgentsCurrentKind {
	currentNAs := &common.NodeAgentsCurrentKind{}
	unmarshalStdoutYaml(orbctlGitops("file", "print", "caos-internal/orbiter/node-agents-current.yml"), currentNAs, false)
	return currentNAs
}

type expectEnsuredOrbiter = func(expectMasters, expectWorkers uint8, k8sVersion string, timeout time.Duration)

func expectEnsuredOrbiterFunc(orbctlGitops orbctlGitopsCmd, kubectl kubectlCmd) expectEnsuredOrbiter {

	return func(expectMasters, expectWorkers uint8, k8sVersion string, timeout time.Duration) {
		print := printOperatorLogs(kubectl)
		type comparable struct {
			orbiterPods, mastersDone, workersDone, nodeAgentsDone uint8
			clusterStatus                                         string
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
				fmt.Printf("node agent %s is not done yet: ready=%t kubeadm=%s kubelet=%s kubectl=%s", naID, na.NodeIsReady, sw.Kubeadm.Version, sw.Kubelet.Version, sw.Kubectl.Version)
			}

			return comparable{
				orbiterPods:    readyPods(kubectl, "caos-system", "app.kubernetes.io/name=orbiter"),
				mastersDone:    mastersDone,
				workersDone:    workersDone,
				clusterStatus:  currentOrbiter.Status,
				nodeAgentsDone: nodeAgentsDone,
			}
		}, timeout, 5).Should(Equal(comparable{
			orbiterPods:    1,
			mastersDone:    expectMasters,
			workersDone:    expectWorkers,
			clusterStatus:  "running",
			nodeAgentsDone: expectMasters + expectWorkers,
		}))
	}
}

type expectUpdatedOrbiter func(patchPath, patchValue, expectK8sVersion string, expectMasters, expectWorkers uint8, timeout time.Duration)

func expectUpdatedOrbiterFunc(orbctlGitops orbctlGitopsCmd, ExpectEnsuredOrbiter expectEnsuredOrbiter) expectUpdatedOrbiter {
	return func(patchPath, patchValue, expectK8sVersion string, expectMasters, expectWorkers uint8, timeout time.Duration) {

		By(fmt.Sprintf("patching the orbiter.yml at %s using the value %s", patchPath, patchValue))

		session, err := gexec.Start(orbctlGitops("file", "patch", "orbiter.yml", patchPath, "--value", patchValue, "--exact"), GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 1*time.Minute).Should(gexec.Exit(0))

		By("waiting for ORBITER to ensure the result")

		ExpectEnsuredOrbiter(expectMasters, expectWorkers, expectK8sVersion, timeout)
	}
}

func someMaster(orbctlGitops orbctlGitopsCmd) (context string, id string) {

	context = "providerundertest.management"
	session, err := gexec.Start(orbctlGitops("nodes", "list", "--context", context, "--column", "id"), GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 1*time.Minute).Should(gexec.Exit(0))

	return context, strings.Split(string(session.Out.Contents()), "\n")[0]
}
