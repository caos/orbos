package main

import (
	"flag"
	"fmt"
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

	flag.Parse()

	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		panic(err)
	}

	original := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(string(out)), "refs/"), "heads/")

	testFunc := run
	/*
		if ghToken != "" {
			testFunc = func(branch string) error {
				return github(trimBranch(branch), ghToken, strings.ToLower(testcase), run)(orbconfig)
			}
		}
	*/
	if graphiteURL != "" {

		orb, err := orb.ParseOrbConfig(helpers.PruneHome(orbconfig))
		if err != nil {
			panic(err)
		}

		testFunc = func(branch, orbconfig string) error {
			branch = strings.ReplaceAll(strings.TrimPrefix(branch, "origin/"), ".", "-")
			return graphite(
				strings.ToLower(strings.Split(strings.Split(orb.URL, "/")[1], ".")[0]),
				graphiteURL,
				graphiteKey,
				trimBranch(branch),
				run)(branch, orbconfig)
		}
	}

	if !unpublished {
		if err := testFunc(original, orbconfig); err != nil {
			panic(err)
		}
		return
	}

	defer func() {
		r := recover()
		if err := checkout(original); err != nil {
			panic(fmt.Errorf("checking out original branch failed: %w: original error: %v", err, r))
		}
		if r != nil {
			panic(r)
		}
	}()

	out, err = exec.Command("git", "branch", "-r", "--no-merged", "origin/master").Output()
	if err != nil {
		panic(err)
	}

	for _, ref := range strings.Fields(string(out)) {
		if checkoutErr := checkout(ref); checkoutErr != nil {
			panic(checkoutErr)
		}
		err = helpers.Concat(err, testFunc(ref, orbconfig))
	}
	if err != nil {
		panic(err)
	}
}

func trimBranch(ref string) string {
	return strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(ref), "refs/"), "heads/")
}

func checkout(ref string) error {
	out, err := exec.Command("git", "checkout", ref).CombinedOutput()
	fmt.Printf(string(out))
	if err != nil {
		return fmt.Errorf("checking out %s failed: %w", ref, err)
	}
	return err
}
