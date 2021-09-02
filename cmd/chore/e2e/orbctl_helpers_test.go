package e2e_test

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/caos/orbos/cmd/chore/e2e"
	"github.com/caos/orbos/internal/operator/common"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v3"

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

func parseBool(val string) bool {
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
	cmdFunc, error := e2e.Command(false, true, false, "")
	Expect(error).ToNot(HaveOccurred())
	return func(args ...string) *exec.Cmd {
		cmd := cmdFunc(context.Background())
		cmd.Args = append(cmd.Args, append([]string{"--disable-analytics", "--gitops", "--orbconfig", orbconfig}, args...)...)
		return cmd
	}
}

func unmarshale2eYmlFunc(orbctl orbctlGitopsCmd) func(interface{}) {

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

func kubectlCmdFunc(kubectlPath string, orbctl orbctlGitopsCmd) kubectlCmd {
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

func mustUnmarshalStdoutYaml(cmd *exec.Cmd, into interface{}) {

	session, err := gexec.Start(cmd, nil, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 2*time.Minute).Should(gexec.Exit(0))

	unmarshalYaml(session.Out.Contents(), into)
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

	mustUnmarshalStdoutYaml(orbctlGitops("file", "print", "caos-internal/orbiter/current.yml"), &currentOrbiter)
	return currentOrbiter
}

func currentNodeagents(orbctlGitops orbctlGitopsCmd) *common.NodeAgentsCurrentKind {
	currentNAs := &common.NodeAgentsCurrentKind{}
	mustUnmarshalStdoutYaml(orbctlGitops("file", "print", "caos-internal/orbiter/node-agents-current.yml"), currentNAs)
	return currentNAs
}

func someMaster(orbctlGitops orbctlGitopsCmd) (context string, id string) {

	context = "providerundertest.management"
	session, err := gexec.Start(orbctlGitops("nodes", "list", "--context", context, "--column", "id"), GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 1*time.Minute).Should(gexec.Exit(0))

	return context, strings.Split(string(session.Out.Contents()), "\n")[0]
}
