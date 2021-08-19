package main

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/caos/orbos/cmd/chore/orbctl"
	"github.com/pkg/errors"
)

func main() {

	var (
		debug, reuse, download bool
		downloadTag            string
	)

	var cmdArgs []string
	var skip bool
	for idx, arg := range os.Args {
		if skip {
			skip = false
			continue
		}
		if strings.HasPrefix(arg, "--bin-download-tag=") {
			downloadTag = strings.TrimPrefix(arg, "--bin-download-tag=")
			continue
		}
		switch arg {
		case "--bin-debug":
			debug = true
			continue
		case "--bin-reuse":
			reuse = true
			continue
		case "--bin-download":
			download = true
			continue
		case "--bin-download-tag":
			downloadTag = os.Args[idx+1]
			skip = true
			continue
		}
		cmdArgs = append(cmdArgs, arg)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	go func() {
		<-signalChannel
		cancel()
	}()

	newOrbctl, err := orbctl.Command(debug, reuse, download, downloadTag)
	if err != nil {
		panic(err)
	}

	cmd := newOrbctl(ctx)
	cmd.Args = append(cmd.Args, cmdArgs[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(err.(*exec.ExitError).ExitCode())
		}
		panic(err)
	}
}
