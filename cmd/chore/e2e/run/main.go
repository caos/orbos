package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/caos/orbos/internal/helpers"
)

func main() {

	var (
		unpublished bool
		orbconfig   string
		ghToken     string
		testcase    string
	)

	const (
		unpublishedDefault  = false
		unpublishedUsage    = "Test all unpublished branches"
		orbDefault          = "~/.orb/config"
		orbUsage            = "Path to the orbconfig file which points to the orb the end-to-end testing should be performed on"
		githubTokenDefault  = ""
		githubTokenKeyUsage = "Personal access token with repo scope for github.com/caos/orbos"
		testcaseDefault     = ""
		testcaseUsage       = "Testcase identifier"
	)

	flag.BoolVar(&unpublished, "unpublished", unpublishedDefault, unpublishedUsage)
	flag.BoolVar(&unpublished, "u", unpublishedDefault, unpublishedUsage+" (shorthand)")
	flag.StringVar(&orbconfig, "orbconfig", orbDefault, orbUsage)
	flag.StringVar(&orbconfig, "f", orbDefault, orbUsage+" (shorthand)")
	flag.StringVar(&ghToken, "github-access-token", githubTokenDefault, githubTokenKeyUsage)
	flag.StringVar(&ghToken, "t", githubTokenDefault, githubTokenKeyUsage+" (shorthand)")
	flag.StringVar(&testcase, "testcase", testcaseDefault, testcaseUsage)
	flag.StringVar(&testcase, "c", testcaseDefault, testcaseUsage+" (shorthand)")

	flag.Parse()

	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		panic(err)
	}

	original := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(string(out)), "refs/"), "heads/")

	testFunc := func(_ string) error {
		return run(orbconfig)
	}

	if ghToken != "" {
		testFunc = func(branch string) error {
			return github(trimBranch(branch), ghToken, strings.ToLower(testcase), run)(orbconfig)
		}
	}

	if !unpublished {
		if err := testFunc(original); err != nil {
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
		err = helpers.Concat(err, testFunc(ref))
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
