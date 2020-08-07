package main

import (
	"flag"
	"os/exec"
	"strings"

	"github.com/caos/orbos/cmd/chore/e2e/shared"
)

func main() {

	var token, org, repository, testcase string

	flag.StringVar(&token, "access-token", "", "Personal access token with repo scope")
	flag.StringVar(&org, "organization", "", "Github organization")
	flag.StringVar(&repository, "repository", "", "Github project")
	flag.StringVar(&testcase, "testcase", "unknown", "Testcase identifier")

	flag.Parse()

	ref, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		panic(err)
	}

	if err := shared.Emit(shared.Event{
		EventType: "webhook-trigger",
		ClientPayload: map[string]string{
			"branch":   strings.TrimPrefix(strings.TrimSpace(string(ref)), "heads/"),
			"testcase": strings.ToLower(testcase),
		},
	}, token, org, repository); err != nil {
		panic(err)
	}
}
