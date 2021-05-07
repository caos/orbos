package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/caos/orbos/pkg/orb"

	"github.com/afiskon/promtail-client/promtail"
	"github.com/caos/orbos/internal/helpers"
)

func main() {

	var (
		unpublished bool
		orbconfig   string
		//		ghToken     string
		graphiteURL string
		graphiteKey string
		lokiURL     string
		from        int
		cleanup     bool
	)

	const (
		unpublishedDefault = false
		unpublishedUsage   = "Test all unpublished branches"
		orbDefault         = "~/.orb/config"
		orbUsage           = "Path to the orbconfig file which points to the orb the end-to-end testing should be performed on"
		//		githubTokenDefault  = ""
		//		githubTokenKeyUsage = "Personal access token with repo scope for github.com/caos/orbos"
		graphiteURLDefault = ""
		graphiteURLUsage   = "https://<your-subdomain>.hosted-metrics.grafana.net/metrics"
		graphiteKeyDefault = ""
		graphiteKeyUsage   = "your api key from grafana.net -- should be editor role"
		lokiURLDefault     = ""
		lokiURLUsage       = "https://<instanceId>:<apiKey>@<instanceUrl>/api/prom/push"
		fromDefault        = 1
		fromUsage          = "step to continue e2e tests from"
		cleanupDefault     = true
		cleanupUsage       = "destroy orb after tests are done"
	)

	flag.BoolVar(&unpublished, "unpublished", unpublishedDefault, unpublishedUsage)
	flag.BoolVar(&unpublished, "u", unpublishedDefault, unpublishedUsage+" (shorthand)")
	flag.StringVar(&orbconfig, "orbconfig", orbDefault, orbUsage)
	flag.StringVar(&orbconfig, "f", orbDefault, orbUsage+" (shorthand)")
	//	flag.StringVar(&ghToken, "github-access-token", githubTokenDefault, githubTokenKeyUsage)
	//	flag.StringVar(&ghToken, "t", githubTokenDefault, githubTokenKeyUsage+" (shorthand)")
	flag.StringVar(&graphiteURL, "graphiteurl", graphiteURLDefault, graphiteURLUsage)
	flag.StringVar(&graphiteURL, "g", graphiteURLDefault, graphiteURLUsage+" (shorthand)")
	flag.StringVar(&graphiteKey, "graphitekey", graphiteKeyDefault, graphiteKeyUsage)
	flag.StringVar(&graphiteKey, "k", graphiteKeyDefault, graphiteKeyUsage+" (shorthand)")
	flag.StringVar(&lokiURL, "lokiurl", lokiURLDefault, lokiURLUsage)
	flag.StringVar(&lokiURL, "l", lokiURLDefault, lokiURLUsage+" (shorthand)")
	flag.BoolVar(&cleanup, "cleanup", cleanupDefault, cleanupUsage)
	flag.BoolVar(&cleanup, "c", cleanupDefault, cleanupUsage+" (shorthand)")
	flag.IntVar(&from, "from", fromDefault, fromUsage)
	flag.IntVar(&from, "s", fromDefault, fromUsage)

	flag.Parse()

	fmt.Printf("unpublished=%t\n", unpublished)
	fmt.Printf("orbconfig=%s\n", orbconfig)
	fmt.Printf("graphiteurl=%s\n", graphiteURL)
	fmt.Printf("cleanup=%t\n", cleanup)
	fmt.Printf("from=%d\n", from)

	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		panic(err)
	}

	branch := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(string(out)), "refs/"), "heads/")
	fmt.Printf("branch=%s\n", branch)

	orbCfg, err := orb.ParseOrbConfig(helpers.PruneHome(orbconfig))
	if err != nil {
		panic(err)
	}

	if err := orb.IsComplete(orbCfg); err != nil {
		panic(err)
	}

	orb := strings.ToLower(strings.Split(strings.Split(orbCfg.URL, "/")[1], ".")[0])

	sendLevel := promtail.DISABLE
	if lokiURL != "" {
		fmt.Println("Sending logs to Loki")
		sendLevel = promtail.INFO
	}

	logger, err := promtail.NewClientProto(promtail.ClientConfig{
		PushURL:            lokiURL,
		Labels:             fmt.Sprintf(`{e2e_test="true", branch="%s", orb="%s"}`, branch, orb),
		BatchWait:          1 * time.Second,
		BatchEntriesNumber: 0,
		SendLevel:          sendLevel,
		PrintLevel:         promtail.DEBUG,
	})
	if err != nil {
		panic(err)
	}
	defer logger.Shutdown()

	testFunc := runFunc
	/*
		if ghToken != "" {
			testFunc = func(branch string) error {
				return github(trimBranch(branch), ghToken, strings.ToLower(testcase), runFunc)(orbconfig)
			}
		}
	*/
	if graphiteURL != "" {
		fmt.Println("Sending status to Graphite")
		testFunc = graphite(
			orb,
			graphiteURL,
			graphiteKey,
			trimBranch(branch),
			runFunc)
	}

	if err := testFunc(logger, strings.ReplaceAll(strings.TrimPrefix(branch, "origin/"), ".", "-"), orbconfig, from, cleanup)(); err != nil {
		logger.Errorf("End-to-end test failed: %w", err)
		panic(err)
	}
	return
}

func trimBranch(ref string) string {
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(ref), "refs/"), "heads/")
}
