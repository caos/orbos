package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"

	"github.com/caos/orbos/cmd/chore/e2e/shared"
)

func main() {

	var token, org, repository, testcase, branch, from string

	flag.StringVar(&token, "access-token", "", "Personal access token with repo scope")
	flag.StringVar(&org, "organization", "", "Github organization")
	flag.StringVar(&repository, "repository", "", "Github project")
	flag.StringVar(&testcase, "testcase", "unknown", "Testcase identifier")
	flag.StringVar(&branch, "branch", "", "Branch to test. Default is current")
	flag.StringVar(&from, "from", "", "From e2e test stage")

	flag.Parse()

	if branch == "" {
		ref, err := exec.Command("git", "branch", "--show-current").Output()
		if err != nil {
			panic(err)
		}
		fmt.Print("Current Branch: ", string(ref))

		branch = strings.TrimPrefix(strings.TrimSpace(string(ref)), "heads/")
	}

	fmt.Println("Tested branch:", branch)
	if err := shared.Emit(shared.Event{
		EventType: "webhook-trigger",
		ClientPayload: map[string]string{
			"branch":   branch,
			"testcase": strings.ToLower(testcase),
			"from":     from,
		},
	}, token, org, repository); err != nil {
		panic(err)
	}
}
