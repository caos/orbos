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
		graphiteURL string
		graphiteKey string
	)

	const (
		unpublishedDefault = false
		unpublishedUsage   = "Test all unpublished branches"
		orbDefault         = "~/.orb/config"
		orbUsage           = "Path to the orbconfig file which points to the orb the end-to-end testing should be performed on"
		graphiteURLDefault = ""
		graphiteURLUsage   = "https://<your-subdomain>.hosted-metrics.grafana.net/metrics"
		graphiteKeyDefault = ""
		graphiteKeyUsage   = "your api key from grafana.net -- should be editor role"
	)

	flag.BoolVar(&unpublished, "unpublished", unpublishedDefault, unpublishedUsage)
	flag.BoolVar(&unpublished, "u", unpublishedDefault, unpublishedUsage+" (shorthand)")
	flag.StringVar(&orbconfig, "orbconfig", orbDefault, orbUsage)
	flag.StringVar(&orbconfig, "f", orbDefault, orbUsage+" (shorthand)")
	flag.StringVar(&graphiteURL, "graphiteurl", graphiteURLDefault, graphiteURLUsage)
	flag.StringVar(&graphiteURL, "g", graphiteURLDefault, graphiteURLUsage+" (shorthand)")
	flag.StringVar(&graphiteKey, "graphitekey", graphiteKeyDefault, graphiteKeyUsage)
	flag.StringVar(&graphiteKey, "k", graphiteKeyDefault, graphiteKeyUsage+" (shorthand)")

	flag.Parse()

	orb, err := orb.ParseOrbConfig(helpers.PruneHome(orbconfig))
	if err != nil {
		panic(err)
	}

	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		panic(err)
	}

	original := strings.TrimSpace(string(out))

	testFunc := run
	if graphiteURL != "" {
		testFunc = graphite(
			strings.ToLower(strings.ReplaceAll(strings.Split(strings.Split(orb.URL, "/")[1], ".")[0], "-", "")),
			graphiteURL,
			graphiteKey,
			strings.Join(strings.Split(original, "/")[1:], "/"),
			run)
	}

	if !unpublished {
		if err := testFunc(orbconfig); err != nil {
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

	out, err = exec.Command("git", "for-each-ref", "--sort", "creatordate", "--format", "%(refname)", "refs/tags", "--no-merged").Output()
	if err != nil {
		panic(err)
	}

	for _, ref := range strings.Fields(string(out)) {
		if checkoutErr := checkout(ref); checkoutErr != nil {
			panic(checkoutErr)
		}
		err = helpers.Concat(err, testFunc(orbconfig))
	}
	if err != nil {
		panic(err)
	}
}

func checkout(ref string) error {
	out, err := exec.Command("git", "checkout", strings.TrimSpace(ref)).CombinedOutput()
	fmt.Printf(string(out))
	if err != nil {
		return fmt.Errorf("checking out %s failed: %w", ref, err)
	}
	return err
}
