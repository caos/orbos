package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/caos/orbos/internal/helpers"
)

func main() {

	var (
		unpublished bool
		orbconfig   string
	)

	const (
		unpublishedDefault = false
		unpublishedUsage   = "Test all unpublished branches"
		orbDefault         = "~/.orb/config"
		orbUsage           = "Path to the orbconfig file which points to the orb the end-to-end testing should be performed on"
	)

	flag.BoolVar(&unpublished, "unpublished", unpublishedDefault, unpublishedUsage)
	flag.BoolVar(&unpublished, "u", unpublishedDefault, unpublishedUsage+" (shorthand)")
	flag.StringVar(&orbconfig, "orbconfig", orbDefault, orbUsage)
	flag.StringVar(&orbconfig, "f", orbDefault, orbUsage+" (shorthand)")

	flag.Parse()

	if !unpublished {
		if err := testCurrentCommit(orbconfig); err != nil {
			panic(err)
		}
		return
	}

	original, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	defer func() {
		r := recover()
		if err := checkout(string(original)); err != nil {
			panic(fmt.Errorf("checking out original branch failed: %w: original error: %v", err, r))
		}
		if r != nil {
			panic(r)
		}
	}()

	out, err := exec.Command("git", "for-each-ref", "--sort", "creatordate", "--format", "%(refname)", "refs/tags", "--no-merged").Output()
	if err != nil {
		panic(err)
	}

	for _, ref := range strings.Fields(string(out)) {
		if checkoutErr := checkout(ref); checkoutErr != nil {
			panic(checkoutErr)
		}
		if testErr := testCurrentCommit(orbconfig); testErr != nil {
			helpers.Concat(err, testErr)
		}
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

func testCurrentCommit(orbconfig string) error {
	files, err := filepath.Glob("./cmd/chore/orbctl/*.go")
	if err != nil {
		panic(err)
	}

	args := []string{"run"}
	args = append(args, files...)
	args = append(args, "--orbconfig", orbconfig)
	args = append(args, "destroy")

	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		if strings.HasPrefix(line, "Are you absolutely sure") {
			if _, err := stdin.Write([]byte("y\n")); err != nil {
				panic(err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
	return nil
}
