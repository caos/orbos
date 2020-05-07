package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator"
	"github.com/caos/orbos/internal/watcher/cron"
	"github.com/caos/orbos/internal/watcher/immediate"
	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/internal/operator/nodeagent/dep"
	"github.com/caos/orbos/internal/operator/nodeagent/dep/conv"
	"github.com/caos/orbos/internal/operator/nodeagent/firewall"
)

var gitCommit string
var version string

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {

	defer func() {
		r := recover()
		if r != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("\x1b[0;31m%v\x1b[0m\n", r)))
			os.Exit(1)
		}
	}()

	verbose := flag.Bool("verbose", false, "Print logs for debugging")
	printVersion := flag.Bool("version", false, "Print build information")
	repoURL := flag.String("repourl", "", "Repository URL")
	ignorePorts := flag.String("ignore-ports", "", "Comma separated list of firewall ports that are ignored")
	nodeAgentID := flag.String("id", "", "The managed machines ID")

	flag.Parse()

	if *printVersion {
		fmt.Println(fmt.Sprintf("%s %s", version, gitCommit))
		os.Exit(0)
	}

	if *repoURL == "" || *nodeAgentID == "" {
		panic("flags --repourl and --id are required")
	}
	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	if *verbose {
		monitor = monitor.Verbose()
	}

	monitor.WithFields(map[string]interface{}{
		"version":     version,
		"commit":      gitCommit,
		"verbose":     *verbose,
		"repourl":     *repoURL,
		"nodeAgentID": *nodeAgentID,
	}).Info("Node Agent is starting")

	os, err := dep.GetOperatingSystem()
	if err != nil {
		panic(err)
	}

	repoKeyPath := "/etc/nodeagent/repokey"
	repoKey, err := ioutil.ReadFile(repoKeyPath)
	if err != nil {
		panic(fmt.Sprintf("repokey not found at %s", repoKeyPath))
	}

	pruned := strings.Split(string(repoKey), "-----")[2]
	hashed := sha256.Sum256([]byte(pruned))
	conv := conv.New(monitor, os, fmt.Sprintf("%x", hashed[:]))

	ctx := context.Background()
	gitClient := git.New(ctx, monitor, fmt.Sprintf("Node Agent %s", *nodeAgentID), "node-agent@caos.ch", *repoURL)
	if err := gitClient.Init(repoKey); err != nil {
		panic(err)
	}

	op := operator.New(
		ctx,
		monitor,
		nodeagent.Iterator(
			monitor,
			gitClient,
			gitCommit,
			*nodeAgentID,
			firewall.Ensurer(monitor, os.OperatingSystem, strings.Split(*ignorePorts, ",")),
			conv,
			conv.Init()),
		[]operator.Watcher{
			immediate.New(monitor),
			cron.New(monitor, "@every 10s"),
		})

	if err := op.Initialize(); err != nil {
		panic(err)
	}
	op.Run()
}
