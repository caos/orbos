package e2e_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/caos/orbos/cmd/chore/e2e"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("orbctl", func() {

	const (
		envPrefix  = "ORBOS_E2E_"
		tagEnv     = envPrefix + "TAG"
		orbEnv     = envPrefix + "ORBURL"
		ghpatEnv   = envPrefix + "GITHUB_ACCESS_TOKEN"
		cleanupEnv = envPrefix + "CLEANUP"
	)

	var (
		tag, orbURL, workfolder, orbconfig, ghTokenPath, accessToken string
		cleanup                                                      bool
		orbctlGitops                                                 orbctlGitopsCmd
		kubectl                                                      kubectlCmd
		e2eYml                                                       func(into interface{})
		AwaitEnsuredORBOS                                            awaitEnsuredORBOS
		AwaitUpdatedOrbiter                                          awaitUpdatedOrbiter
	)

	BeforeSuite(func() {
		workfolder = "./artifacts"
		orbconfig = filepath.Join(workfolder, "orbconfig")
		ghTokenPath = filepath.Join(workfolder, "ghtoken")
		tag = os.Getenv(tagEnv)
		orbURL = os.Getenv(orbEnv)
		accessToken = os.Getenv(ghpatEnv)
		cleanup = parseBool(os.Getenv(cleanupEnv))
		orbctlGitops = orbctlGitopsFunc(orbconfig)
		e2eYml = unmarshale2eYmlFunc(orbctlGitops)
		kubectl = kubectlCmdFunc(filepath.Join(workfolder, "kubeconfig"), orbctlGitops)
		AwaitEnsuredORBOS = awaitEnsuredOrbiterFunc(orbctlGitops, kubectl)
		AwaitUpdatedOrbiter = awaitUpdatedOrbiterFunc(orbctlGitops, AwaitEnsuredORBOS)

		Expect(tag).ToNot(BeEmpty(), fmt.Sprintf("environment variable %s is required", tagEnv))
		Expect(orbURL).ToNot(BeEmpty(), fmt.Sprintf("environment variable %s is required", orbEnv))
		Expect(accessToken).ToNot(BeEmpty(), fmt.Sprintf("environment variable %s is required", ghpatEnv))
	})

	AfterSuite(func() {
		if !cleanup {
			return
		}
		cmd := orbctlGitops("destroy")
		cmd.Stdin = strings.NewReader("yes")

		session, err := gexec.Start(cmd, os.Stdout, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 5*time.Minute).Should(gexec.Exit(0))
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

				fmt.Println("DEBUG: download orbctl")

				cmdFunc, err := e2e.Command(false, false, true, tag)
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
			/*			It("truncates the orbconfig", func() {
						orbconfigFile, err := os.Create(orbconfig)
						Expect(err).ToNot(HaveOccurred())
						defer orbconfigFile.Close()
					})*/

			It("ensures the ghtoken cache file so that the oauth flow is skipped", func() {
				fmt.Println("DEBUG: write local files")

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

				fmt.Println("DEBUG: create orbconfig")

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

				fmt.Println("DEBUG: create remote boom.yml")

				It("succeeds when creating the initial boom.yml", func() {
					contentBytes, err := ioutil.ReadFile("./boom-init.yml")
					Expect(err).ToNot(HaveOccurred())

					session, err := gexec.Start(orbctlGitops("file", "patch", "boom.yml", "--exact", "--value", os.ExpandEnv(string(contentBytes))), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(session, 1*time.Minute).Should(gexec.Exit(0))
				})

				It("succeeds when creating the initial orbiter.yml", func() {

					fmt.Println("DEBUG: create remote orbiter.yml")

					By("fetching the file provider-init.yml from git")

					printSession, err := gexec.Start(orbctlGitops("file", "print", "provider-init.yml"), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(printSession, 1*time.Minute).Should(gexec.Exit(0))
					providerSpecs := "    " + strings.Join(strings.Split(string(printSession.Out.Contents()), "\n"), "\n    ")

					By("reading the file ./orbiter-init.yml from the file system")

					contentBytes, err := ioutil.ReadFile("./orbiter-init.yml")
					Expect(err).ToNot(HaveOccurred())

					orbiterYml := os.ExpandEnv(fmt.Sprintf(string(contentBytes), providerSpecs))

					By("replacing the file orbiter.yml in git")

					patchSession, err := gexec.Start(orbctlGitops("file", "patch", "orbiter.yml", "--exact", "--value", orbiterYml), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(patchSession, 1*time.Minute).Should(gexec.Exit(0))
				})

				It("migrates the api successfully", func() {

					fmt.Println("DEBUG: migrate")

					patchSession, err := gexec.Start(orbctlGitops("api"), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(patchSession, 1*time.Minute).Should(gexec.Exit(0))
				})

				It("creates provider specific secrets successfully", func() {

					fmt.Println("DEBUG: write provider secrets")

					cfg := struct {
						Initsecrets map[string]string
					}{}
					e2eYml(&cfg)
					Expect(cfg.Initsecrets).ToNot(HaveLen(0))

					for k, v := range cfg.Initsecrets {
						secretKey := fmt.Sprintf("orbiter.providerundertest.%s.encrypted", k)

						By(fmt.Sprintf("writing the secret %s using the value from environment variable %s", secretKey, v))

						expanded := os.Getenv(v)
						Expect(expanded).ToNot(BeEmpty())
						session, err := gexec.Start(orbctlGitops("writesecret", secretKey, "--value", expanded), GinkgoWriter, GinkgoWriter)
						Expect(err).ToNot(HaveOccurred())
						Eventually(session, 1*time.Minute).Should(gexec.Exit(0))
					}
				})

				It("configures successfully", func() {

					fmt.Println("DEBUG: generate secrets")

					out, err := exec.Command("ls", "-la", "./artifacts").Output()
					Expect(err).ToNot(HaveOccurred())
					fmt.Println("ls -la\n:", string(out))

					orbconfig, err := ioutil.ReadFile("./artifacts/orbconfig")
					Expect(err).ToNot(HaveOccurred())
					fmt.Println("orbconfig\n:", string(orbconfig))

					ghtoken, err := ioutil.ReadFile("./artifacts/ghtoken")
					Expect(err).ToNot(HaveOccurred())
					fmt.Println("orbconfig\n:", string(ghtoken))

					configureSession, err := gexec.Start(orbctlGitops("configure"), GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(configureSession, 2*time.Minute, 1*time.Second).Should(gexec.Exit(0))
				})
			})
		})
	})
	When("bootstrapping", func() {
		It("creates the kubeapi", func() {

			session, err := gexec.Start(orbctlGitops("takeoff"), os.Stdout, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session, 15*time.Minute, 5*time.Second).Should(gexec.Exit(0))

			AwaitEnsuredORBOS(1, 1, "v1.18.8", 10*time.Minute)
		})
	})
	Context("scaling", func() {
		When("desiring a higher workers count", func() {
			It("scales up workers", func() {
				AwaitUpdatedOrbiter("clusters.k8s.spec.workers.0.nodes", "3", "v1.18.8", 1, 3, 10*time.Minute)
			})
		})
		When("desiring a lower workers count", func() {
			It("scales down workers", func() {
				AwaitUpdatedOrbiter("clusters.k8s.spec.workers.0.nodes", "1", "v1.18.8", 1, 1, 10*time.Minute)
			})
		})
		When("desiring a higher masters count", func() {
			It("scales up masters", func() {
				AwaitUpdatedOrbiter("clusters.k8s.spec.controlplane.nodes", "3", "v1.18.8", 3, 1, 15*time.Minute)
			})
		})
		When("desiring a lower masters count", func() {
			It("scales down masters", func() {
				AwaitUpdatedOrbiter("clusters.k8s.spec.controlplane.nodes", "1", "v1.18.8", 1, 1, 10*time.Minute)
			})
		})
	})
	Context("machine", func() {
		When("desiring a machine reboot", func() {
			It("updates the machines last reboot time", func() {

				testStart := time.Now()
				machineContext, machineID := someMaster(orbctlGitops)

				By(fmt.Sprintf("executing the reboot command for machine %s", machineID))

				session, err := gexec.Start(orbctlGitops("node", "reboot", fmt.Sprintf("%s.%s", machineContext, machineID)), GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, 1*time.Minute).Should(gexec.Exit(0))

				By("waiting for ORBITER to ensure the result")

				Eventually(func() time.Time {
					machine, ok := currentNodeagents(orbctlGitops).Current.Get(machineID)
					if !ok {
						return testStart
					}
					return machine.Booted
				}, 5*time.Minute).Should(SatisfyAll(BeTemporally(">", testStart)))

				AwaitEnsuredORBOS(1, 1, "v1.18.8", 5*time.Minute)
			})
		})
		When("desiring a machine replacement", func() {
			It("removes a machine and joins a new one", func() {

				machineContext, machineID := someMaster(orbctlGitops)

				currentMachines := func() map[string]*kubernetes.Machine {
					return currentOrbiter(orbctlGitops).Clusters["k8s"].Current.Machines.M
				}

				var old []string
				for k := range currentMachines() {
					old = append(old, k)
				}
				isNew := func(id string) bool {
					for i := range old {
						if old[i] == id {
							return false
						}
					}
					fmt.Println(id, "is new")
					return true
				}

				By(fmt.Sprintf("executing the replace command for machine %s", machineID))

				session, err := gexec.Start(orbctlGitops("node", "replace", fmt.Sprintf("%s.%s", machineContext, machineID)), GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session, 1*time.Minute).Should(gexec.Exit(0))

				By("waiting for ORBITER to add a new machine")

				Eventually(func() bool {
					fmt.Println("waiting for ORBITER to add a new machine")
					for k := range currentMachines() {
						if isNew(k) {
							return true
						}
					}
					return false
				}, 5*time.Minute).Should(BeTrue())

				By("waiting for ORBITER to ensure the result")

				AwaitEnsuredORBOS(1, 1, "v1.18.8", 15*time.Minute)
				Eventually(func() bool {
					fmt.Println("waiting for ORBITER to remove machine from current state")
					_, ok := currentNodeagents(orbctlGitops).Current.Get(machineID)
					fmt.Println("machine", machineID, "is still listed in current node agents")
					return ok
				}, 2*time.Minute).Should(BeFalse())
			})
		})
	})
	When("desiring the latest kubernetes release", func() {
		It("upgrades the kubernetes binaries", func() {
			AwaitUpdatedOrbiter("clusters.k8s.spec.versions.kubernetes", "v1.21.0", "v1.21.0", 1, 1, 60*time.Minute)
		})
	})
})
