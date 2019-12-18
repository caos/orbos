package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/internal/edge/watcher/cron"
	"github.com/caos/orbiter/internal/edge/watcher/immediate"
	logcontext "github.com/caos/orbiter/logging/context"
	"github.com/caos/orbiter/logging/stdlib"

	"github.com/caos/orbiter/internal/kinds/nodeagent"
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/conv"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/firewall"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/rebooter/node"
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
	computeID := flag.String("id", "", "The managed computes ID")
	configPath := flag.String("yamlbasepath", "", "Point separated yaml path to the node agent kind")
	currentFile := flag.String("currentfile", "", "git path to the yaml file containing the current state")
	secretsFile := flag.String("secretsfile", "", "git path to the yaml file containing secrets")

	flag.Parse()

	if *printVersion {
		fmt.Printf("%s %s\n", version, gitCommit)
		os.Exit(0)
	}

	logger := logcontext.Add(stdlib.New(os.Stderr))
	if *verbose {
		logger = logger.Verbose()
	}
	logger.WithFields(map[string]interface{}{
		"version":    version,
		"commit":     gitCommit,
		"verbose":    *verbose,
		"repourl":    *repoURL,
		"computeId":  *computeID,
		"configPath": *configPath,
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

	converter := conv.New(logger, os, fmt.Sprintf("%x", hashed[:]))
	before, err := converter.Init()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	gitClient := git.New(ctx, logger, fmt.Sprintf("Node Agent %s", *computeID), *repoURL)
	if err := gitClient.Init(repoKey); err != nil {
		panic(err)
	}

	op := operator.New(&operator.Arguments{
		Ctx:         ctx,
		GitClient:   gitClient,
		Logger:      logger,
		CurrentFile: *currentFile,
		SecretsFile: *secretsFile,
		Watchers: []operator.Watcher{
			immediate.New(logger),
			cron.New(logger, "@every 30s"),
		},
		RootAssembler: nodeagent.New(strings.Split(*configPath, "."), nil,
			adapter.New(version, logger, node.New(), firewall.Ensurer(logger, os.OperatingSystem), converter, before, nil)),
	})

	iterations := make(chan *operator.IterationDone)
	if err := op.Initialize(); err != nil {
		panic(err)
	}

	go op.Run(iterations)

	for it := range iterations {
		if it.Error != nil {
			logger.Error(it.Error)
		}
	}
}
