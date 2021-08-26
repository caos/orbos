package orbctl_test

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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/caos/orbos/cmd/chore/orbctl"
)

var _ = Describe("orbctl", func() {

	var (
		tag, orbURL /*, orbID*/, workfolder, orbconfig, ghTokenPath, accessToken string
		orbctlGitops                                                             orbctlGitopsCmd
		kubectl                                                                  kubectlCmd
		e2eYml                                                                   func(into interface{})
		ExpectEnsuredOrbiter                                                     expectEnsuredOrbiter
		ExpectUpdatedOrbiter                                                     expectUpdatedOrbiter
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
		ExpectEnsuredOrbiter = expectEnsuredOrbiterFunc(orbctlGitops, kubectl)
		ExpectUpdatedOrbiter = expectUpdatedOrbiterFunc(orbctlGitops, ExpectEnsuredOrbiter)
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
			/*			It("truncates the orbconfig", func() {
						orbconfigFile, err := os.Create(orbconfig)
						Expect(err).ToNot(HaveOccurred())
						defer orbconfigFile.Close()
					})*/

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

			ExpectEnsuredOrbiter(1, 1, "v1.18.8", 5*time.Minute)
		})
	})
	Context("scaling", func() {
		When("desiring a higher workers count", func() {
			It("scales up workers", func() {
				ExpectUpdatedOrbiter("clusters.k8s.spec.workers.0.nodes", "3", "v1.18.8", 1, 3, 10*time.Minute)
			})
		})
		When("desiring a lower workers count", func() {
			It("scales down workers", func() {
				ExpectUpdatedOrbiter("clusters.k8s.spec.workers.0.nodes", "1", "v1.18.8", 1, 1, 10*time.Minute)
			})
		})
		When("desiring a higher masters count", func() {
			FIt("scales up masters", func() {
				ExpectUpdatedOrbiter("clusters.k8s.spec.controlplane.nodes", "3", "v1.18.8", 3, 1, 10*time.Minute)
			})
		})
		When("desiring a lower masters count", func() {
			It("scales down masters", func() {
				ExpectUpdatedOrbiter("clusters.k8s.spec.controlplane.nodes", "1", "v1.18.8", 1, 1, 10*time.Minute)
			})
		})
	})
	Context("machine", func() {
		When("desiring a machine reboot", func() {
			It("updates the machines last reboot time", func() {

			})
		})
		When("desiring a machine replacement", func() {
			It("removes a machine and joins a new one", func() {

			})
		})
	})
	Context("kubernetes upgrading", func() {
		When("desiring the latest kubernetes release", func() {
			It("upgrades the kubernetes binaries", func() {
				ExpectUpdatedOrbiter("clusters.k8s.spec.versions.kubernetes", "v1.21.0", "v1.21.0", 1, 1, 10*time.Minute)
			})
		})
	})
})
