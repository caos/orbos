package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/caos/orbos/cmd/chore"
)

func main() {

	var debug bool
	for idx, arg := range os.Args {
		if arg == "--debug" {
			debug = true
			os.Args = append(os.Args[0:idx], os.Args[idx+1:]...)
			break
		}
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	outBuf := new(bytes.Buffer)
	cmd.Stdout = outBuf
	chore.Run(cmd)

	version := strings.TrimSpace(strings.Replace(outBuf.String(), "heads/", "", 1))

	cmd = exec.Command("git", "rev-parse", "HEAD")
	outBuf = new(bytes.Buffer)
	cmd.Stdout = outBuf
	chore.Run(cmd)

	commit := strings.TrimSpace(outBuf.String())

	files, err := filepath.Glob("./cmd/chore/gen-executables/*.go")
	if err != nil {
		panic(err)
	}
	args := []string{"run", "-race"}
	args = append(args, files...)
	args = append(args,
		"--version", version,
		"--commit", commit,
		"--githubclientid", os.Getenv("ORBOS_GITHUBOAUTHCLIENTID"),
		"--githubclientsecret", os.Getenv("ORBOS_GITHUBOAUTHCLIENTSECRET"),
		"--orbctl", "./artifacts",
		"--dev",
	)
	if debug {
		args = append(args, "--debug")
	}
	cmd = exec.Command("go", args...)
	cmd.Stdout = os.Stderr
	chore.Run(cmd)

	files, err = filepath.Glob("./cmd/chore/gen-charts/*.go")
	if err != nil {
		panic(err)
	}
	args = []string{"build", "-o", "./artifacts/gen-charts"}
	args = append(args, files...)
	cmd = exec.Command("go", args...)
	cmd.Stdout = os.Stderr
	cmd.Env = []string{"CGO_ENABLED=0", "GOOS=linux"}
	chore.Run(cmd)

	if debug {
		args = []string{"exec", "--api-version", "2", "--headless", "--listen", "127.0.0.1:2345", "./artifacts/orbctl-Linux-x86_64", "--"}
		args = append(args, os.Args[1:]...)
		cmd = exec.Command("dlv", args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		return
	}

	cmd = exec.Command("./artifacts/orbctl-Linux-x86_64", os.Args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
