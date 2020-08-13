package main

import (
	"flag"
	"os/exec"
	"strings"

	"github.com/caos/orbos/internal/orb"

	"github.com/caos/orbos/internal/helpers"
)

func main() {

	var (
		unpublished bool
		orbconfig   string
		//		ghToken     string
		//		testcase    string
		graphiteURL string
		graphiteKey string
		from        int
	)

	const (
		unpublishedDefault = false
		unpublishedUsage   = "Test all unpublished branches"
		orbDefault         = "~/.orb/config"
		orbUsage           = "Path to the orbconfig file which points to the orb the end-to-end testing should be performed on"
		//		githubTokenDefault  = ""
		//		githubTokenKeyUsage = "Personal access token with repo scope for github.com/caos/orbos"
		//		testcaseDefault     = ""
		//		testcaseUsage       = "Testcase identifier"
		graphiteURLDefault = ""
		graphiteURLUsage   = "https://<your-subdomain>.hosted-metrics.grafana.net/metrics"
		graphiteKeyDefault = ""
		graphiteKeyUsage   = "your api key from grafana.net -- should be editor role"
		fromDefault        = 1
		fromUsage          = "step to continue e2e tests from"
	)

	flag.BoolVar(&unpublished, "unpublished", unpublishedDefault, unpublishedUsage)
	flag.BoolVar(&unpublished, "u", unpublishedDefault, unpublishedUsage+" (shorthand)")
	flag.StringVar(&orbconfig, "orbconfig", orbDefault, orbUsage)
	flag.StringVar(&orbconfig, "f", orbDefault, orbUsage+" (shorthand)")
	//	flag.StringVar(&ghToken, "github-access-token", githubTokenDefault, githubTokenKeyUsage)
	//	flag.StringVar(&ghToken, "t", githubTokenDefault, githubTokenKeyUsage+" (shorthand)")
	//	flag.StringVar(&testcase, "testcase", testcaseDefault, testcaseUsage)
	//	flag.StringVar(&testcase, "c", testcaseDefault, testcaseUsage+" (shorthand)")
	flag.StringVar(&graphiteURL, "graphiteurl", graphiteURLDefault, graphiteURLUsage)
	flag.StringVar(&graphiteURL, "g", graphiteURLDefault, graphiteURLUsage+" (shorthand)")
	flag.StringVar(&graphiteKey, "graphitekey", graphiteKeyDefault, graphiteKeyUsage)
	flag.StringVar(&graphiteKey, "k", graphiteKeyDefault, graphiteKeyUsage+" (shorthand)")
	flag.IntVar(&from, "from", fromDefault, fromUsage)
	flag.IntVar(&from, "s", fromDefault, fromUsage)

	flag.Parse()

	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		panic(err)
	}

	branch := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(string(out)), "refs/"), "heads/")

	testFunc := runFunc
	/*
		if ghToken != "" {
			testFunc = func(branch string) error {
				return github(trimBranch(branch), ghToken, strings.ToLower(testcase), runFunc)(orbconfig)
			}
		}
	*/
	if graphiteURL != "" {

		orb, err := orb.ParseOrbConfig(helpers.PruneHome(orbconfig))
		if err != nil {
			panic(err)
		}

		if err := orb.IsComplete(); err != nil {
			panic(err)
		}

		testFunc = graphite(
			strings.ToLower(strings.Split(strings.Split(orb.URL, "/")[1], ".")[0]),
			graphiteURL,
			graphiteKey,
			trimBranch(branch),
			runFunc)
	}

	if err := testFunc(strings.ReplaceAll(strings.TrimPrefix(branch, "origin/"), ".", "-"), orbconfig, from)(); err != nil {
		panic(err)
	}
	return
}

func trimBranch(ref string) string {
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(ref), "refs/"), "heads/")
}
