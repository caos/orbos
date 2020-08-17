package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/caos/orbos/cmd/chore/e2e/shared"
)

func main() {

	var (
		token, org, repository, testcase, branch string
		from                                     int
		cleanup                                  bool
	)

	flag.StringVar(&token, "access-token", "", "Personal access token with repo scope")
	flag.StringVar(&org, "organization", "", "Github organization")
	flag.StringVar(&repository, "repository", "", "Github project")
	flag.StringVar(&testcase, "testcase", "unknown", "Testcase identifier")
	flag.StringVar(&branch, "branch", "", "Branch to test. Default is current")
	flag.IntVar(&from, "from", 1, "From e2e test stage")
	flag.BoolVar(&cleanup, "cleanup", true, "Cleanup after tests are done")

	flag.Parse()

	if branch == "" {
		ref, err := exec.Command("git", "branch", "--show-current").Output()
		if err != nil {
			panic(err)
		}
		branch = strings.TrimPrefix(strings.TrimSpace(string(ref)), "heads/")
	}

	fmt.Printf("organization=%s\n", org)
	fmt.Printf("repository=%s\n", repository)
	fmt.Printf("testcase=%s\n", testcase)
	fmt.Printf("branch=%s\n", branch)
	fmt.Printf("from=%d\n", from)
	fmt.Printf("cleanup=%t\n", cleanup)

	if err := shared.Emit(shared.Event{
		EventType: "webhook-trigger",
		ClientPayload: map[string]string{
			"branch":   branch,
			"testcase": strings.ToLower(testcase),
			"from":     strconv.Itoa(from),
			"cleanup":  strconv.FormatBool(cleanup),
		},
	}, token, org, repository); err != nil {
		panic(err)
	}
}
