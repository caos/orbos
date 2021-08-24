package orbctl_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/pkg/orb"

	"github.com/onsi/gomega/gbytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/caos/orbos/cmd/chore/orbctl"
)

var _ = Describe("orbctl", func() {

	var (
		tag, orbURL /*, orbID*/, workfolder, orbconfig, ghTokenPath, accessToken string
		orbctlGitops                                                             orbctlGitopsCmd
		kubectl                                                                  kubectlCmd
		e2eYml                                                                   func(into interface{})
		//		cleanup               bool
	)

	BeforeSuite(func() {
		workfolder = "./artifacts"
		orbconfig = filepath.Join(workfolder, "orbconfig")
		ghTokenPath = filepath.Join(workfolder, "ghtoken")
		tag = prefixedEnv("TAG")
		orbURL = prefixedEnv("ORBURL")
		accessToken = prefixedEnv("GITHUB_ACCESS_TOKEN")
		orbctlGitops = orbctlGitopsFunc(orbconfig)
		e2eYml = memoizeUnmarshalE2eYml(orbctlGitops)
		kubectl = memoizeKubecltCmd(filepath.Join(workfolder, "kubeconfig"), orbctlGitops)
		//		orbID = calcOrbID(orbconfig)
		//		cleanup = boolEnv(prefixedEnv("CLEANUP"))

		Expect(tag).ToNot(BeEmpty())
		Expect(orbconfig).ToNot(BeEmpty())
		Expect(orbURL).ToNot(BeEmpty())
		//		Expect(orbID).ToNot(BeEmpty())
	})

	Context("version", func() {
		/*When("the orbctl is built from source", func() {
			It("contains the current git commit", func() {
				cmdFunc, err := orbctl.Command(false, false, false, "")
				Expect(err).ToNot(HaveOccurred())

				cmd := cmdFunc(context.Background())
				cmd.Args = append(cmd.Args, "--version")

				versionSession, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				commitSession, err := gexec.Start(exec.Command("git", "rev-parse", "HEAD"), GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				Expect(versionSession.Wait()).Should(gbytes.Say(regexp.QuoteMeta(string(commitSession.Wait().Out.Contents()))))
			})
		})*/
		When("the orbctl is downloaded from github releases", func() {
			It("contains the tag read from environment variable", func() {

				cmdFunc, err := orbctl.Command(false, false, true, tag)
				Expect(err).ToNot(HaveOccurred())

				cmd := cmdFunc(context.Background())
				cmd.Args = append(cmd.Args, "--version")

				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, 2*time.Minute, 1*time.Second).Should(gexec.Exit(0))
				Eventually(session, 2*time.Minute, 1*time.Second).Should(gbytes.Say(regexp.QuoteMeta(tag)))
			})
		})
	})

	Context("repository initialization", func() {
		When("initializing local files", func() {
			It("truncates the orbconfig", func() {
				orbconfigFile, err := os.Create(orbconfig)
				Expect(err).ToNot(HaveOccurred())
				defer orbconfigFile.Close()
			})

			It("ensures the ghtoken cache file so that the oauth flow is skipped", func() {
				ghtoken, err := os.Create(ghTokenPath)
				Expect(err).ToNot(HaveOccurred())
				defer ghtoken.Close()
				Expect(ghtoken.WriteString(fmt.Sprintf(`IDToken: ""
IDTokenClaims: null
access_token: %s
expiry: "0001-01-01T00:00:00Z"
token_type: bearer`, accessToken))).To(BeNumerically(">", 0))
			})
		})
		When("configure command is executed for the first time", func() {
			It("creates a new orbconfig containing a new masterkey and a new ssh private key and adds the public key to the repository", func() {
				masterKeySession, err := gexec.Start(exec.Command("openssl", "rand", "-base64", "21"), nil, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(masterKeySession).Should(gexec.Exit(0))

				configureSession, err := gexec.Start(orbctlGitops("configure", "--repourl", orbURL, "--masterkey", string(masterKeySession.Out.Contents())), GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(configureSession, 2*time.Minute, 1*time.Second).Should(gexec.Exit())
			})
		})
		Context("initialized repository access", func() {
			When("creating remote initial files", func() {

				It("succeeds when creating the initial boom.yml", func() {
					contentBytes, err := ioutil.ReadFile("./boom-init.yml")
					Expect(err).ToNot(HaveOccurred())

					session, err := gexec.Start(orbctlGitops("file", "patch", "boom.yml", "--exact", "--value", os.ExpandEnv(string(contentBytes))), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(session, 1*time.Minute).Should(gexec.Exit(0))
				})

				It("succeeds when creating the initial orbiter.yml", func() {
					printSession, err := gexec.Start(orbctlGitops("file", "print", "provider-init.yml"), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(printSession, 1*time.Minute).Should(gexec.Exit(0))
					providerSpecs := "    " + strings.Join(strings.Split(string(printSession.Out.Contents()), "\n"), "\n    ")

					contentBytes, err := ioutil.ReadFile("./orbiter-init.yml")
					Expect(err).ToNot(HaveOccurred())

					orbiterYml := os.ExpandEnv(fmt.Sprintf(string(contentBytes), providerSpecs))

					patchSession, err := gexec.Start(orbctlGitops("file", "patch", "orbiter.yml", "--exact", "--value", orbiterYml), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(patchSession, 1*time.Minute).Should(gexec.Exit(0))
				})

				It("migrates the api successfully", func() {
					patchSession, err := gexec.Start(orbctlGitops("api"), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(patchSession, 1*time.Minute).Should(gexec.Exit(0))
				})

				It("creates provider specific secrets successfully", func() {

					cfg := struct {
						Initsecrets map[string]string
					}{}
					e2eYml(&cfg)
					Expect(cfg.Initsecrets).ToNot(HaveLen(0))

					for k, v := range cfg.Initsecrets {
						expanded := os.Getenv(v)
						Expect(expanded).ToNot(BeEmpty())
						session, err := gexec.Start(orbctlGitops("writesecret", fmt.Sprintf("orbiter.providerundertest.%s.encrypted", k), "--value", expanded), GinkgoWriter, GinkgoWriter)
						Expect(err).ToNot(HaveOccurred())
						Eventually(session, 1*time.Minute).Should(gexec.Exit(0))
					}
				})

				It("configures successfully", func() {
					configureSession, err := gexec.Start(orbctlGitops("configure"), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(configureSession, 2*time.Minute, 1*time.Second).Should(gexec.Exit(0))
				})
			})
		})
	})
	Context("bootstrapping", func() {
		It("creates the kubeapi", func() {

			session, err := gexec.Start(orbctlGitops("takeoff"), os.Stdout, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 15*time.Minute, 5*time.Second).Should(gexec.Exit(0))

		})

		It("ensures the starting position", func() {

			Eventually(readyPods(kubectl, "caos-system", "app.kubernetes.io/name=orbiter"), 10*time.Minute, 5*time.Second).Should(BeIdenticalTo(1))

			/*
				var (
					orbiter    = currentOrbiter{}
					nodeagents = common.NodeAgentsCurrentKind{}
				)

				if err := helpers.Fanout([]func() error{
					func() error {
						return readyPods(ctx, settings, newKubectl, "caos-system", condition.watcher.selector, 1)
					},
					func() error {
						return unmarshalStdoutYaml(ctx, settings, newOrbctl, &orbiter, "--gitops", "file", "print", "caos-internal/orbiter/current.yml")
					},
					func() error {
						return unmarshalStdoutYaml(ctx, settings, newOrbctl, &nodeagents, "--gitops", "file", "print", "caos-internal/orbiter/node-agents-current.yml")
					},
				})(); err != nil {
					return err
				}*/
		})
	})
})

func prefixedEnv(env string) string {
	return os.Getenv("ORBOS_E2E_" + env)
}

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

		session, err := gexec.Start(orbctl("file", "print", "e2e.yml"), GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5*time.Second).Should(gexec.Exit(0))

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

		session, err := gexec.Start(orbctl("readsecret", "orbiter.k8s.kubeconfig.encrypted"), GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5*time.Second).Should(gexec.Exit(0))

		file, err := os.Create(kubectlPath)
		Expect(err).ToNot(HaveOccurred())
		defer file.Close()

		Expect(io.Copy(session.Out, file)).To(BeNumerically("~", 5500, 1000))

		read = true

		return cmd
	}
}

func unmarshalYaml(content []byte, into interface{}) {
	Expect(yaml.Unmarshal(content, into)).To(Succeed())
}

func unmarshalStdoutYaml(cmd *exec.Cmd, into interface{}, timeout time.Duration) {

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, timeout).Should(gexec.Exit(0))

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

	unmarshalStdoutYaml(kubectl(args...), &pods, 5*time.Second)

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
