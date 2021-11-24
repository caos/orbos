package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/caos/orbos/internal/operator/nodeagent/networking"

	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"

	_ "net/http/pprof"

	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/internal/operator/nodeagent/dep"
	"github.com/caos/orbos/internal/operator/nodeagent/dep/conv"
	"github.com/caos/orbos/internal/operator/nodeagent/firewall"
)

var (
	gitCommit string
	version   string
)

func main() {

	ctx, cancelCtx := context.WithCancel(context.Background())

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}

	defer func() { monitor.RecoverPanic(recover()) }()

	verbose := flag.Bool("verbose", false, "Print logs for debugging")
	printVersion := flag.Bool("version", false, "Print build information")
	ignorePorts := flag.String("ignore-ports", "", "Comma separated list of firewall ports that are ignored")
	nodeAgentID := flag.String("id", "", "The managed machines ID")
	pprof := flag.Bool("pprof", false, "start pprof as port 6060")
	sentryEnvironment := flag.String("environment", "", "Sentry environment")

	flag.Parse()

	if *printVersion {
		fmt.Printf("%s %s\n", version, gitCommit)
		os.Exit(0)
	}

	if *sentryEnvironment != "" {
		if err := mntr.Ingest(monitor, "orbos", version, *sentryEnvironment, "node-agent"); err != nil {
			panic(err)
		}
	}

	monitor.WithField("id", nodeAgentID).CaptureMessage("nodeagent invoked")

	if *verbose {
		monitor = monitor.Verbose()
	}

	if *nodeAgentID == "" {
		panic("flag --id is required")
	}

	monitor.WithFields(map[string]interface{}{
		"version":           version,
		"commit":            gitCommit,
		"verbose":           *verbose,
		"nodeAgentID":       *nodeAgentID,
		"sentryEnvironment": *sentryEnvironment,
	}).Info("Node Agent is starting")

	mutexActionChannel := make(chan interface{})

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	go func() {
		for sig := range signalChannel {
<<<<<<< HEAD
			monitor.WithField("signal", sig.String()).Info("Received signal")
			cancelCtx()
=======
>>>>>>> cypress-testing
			mutexActionChannel <- sig
		}
	}()

	if *pprof {
		go func() {
			monitor.Info(http.ListenAndServe("localhost:6060", nil).Error())
		}()
	}

	runningOnOS, err := dep.GetOperatingSystem()
	if err != nil {
		panic(err)
	}

	repoKey, err := nodeagent.RepoKey()
	if err != nil {
		panic(err)
	}

	pruned := strings.Split(string(repoKey), "-----")[2]
	hashed := sha256.Sum256([]byte(pruned))
<<<<<<< HEAD
	conv := conv.New(ctx, monitor, runningOnOS, fmt.Sprintf("%x", hashed[:]))
=======
	conv := conv.New(monitor, runningOnOS, fmt.Sprintf("%x", hashed[:]))
>>>>>>> cypress-testing

	gitClient := git.New(ctx, monitor, fmt.Sprintf("Node Agent %s", *nodeAgentID), "node-agent@caos.ch")

	var portsSlice []string
	if len(*ignorePorts) > 0 {
		portsSlice = strings.Split(*ignorePorts, ",")
	}

	itFunc := nodeagent.Iterator(
		monitor,
		gitClient,
		gitCommit,
		*nodeAgentID,
		firewall.Ensurer(monitor, runningOnOS.OperatingSystem, portsSlice),
		networking.Ensurer(monitor, runningOnOS.OperatingSystem),
		conv,
		conv.Init())

	type updateType struct{}
	go func() {
		for range time.Tick(24 * time.Hour) {
			timer := time.NewTimer(time.Duration(rand.Intn(120)) * time.Minute)
			<-timer.C
			mutexActionChannel <- updateType{}
			timer.Stop()
		}
	}()

	type iterateType struct{}
	//trigger first iteration
	go func() { mutexActionChannel <- iterateType{} }()

	go func() {
		for range time.Tick(5 * time.Minute) {
			if PrintMemUsage(monitor) > 250 {
				monitor.Info("Shutting down as memory usage exceeded 250 MiB")
				mutexActionChannel <- syscall.Signal(0)
			}
		}
	}()

	for action := range mutexActionChannel {
		switch sig := action.(type) {
		case os.Signal:
			monitor.WithField("signal", sig.String()).Info("Shutting down")
<<<<<<< HEAD
			os.Exit(0)
=======
			cancelCtx()
			os.Exit(int(sig.(syscall.Signal)))
>>>>>>> cypress-testing
		case iterateType:
			monitor.Info("Starting iteration")
			itFunc()
			monitor.Info("Iteration done")
			go func() {
				//trigger next iteration
				time.Sleep(10 * time.Second)
				mutexActionChannel <- iterateType{}
			}()
		case updateType:
			monitor.Info("Starting update")
			if err := conv.Update(); err != nil {
				monitor.Error(fmt.Errorf("updating packages failed: %w", err))
			} else {
				monitor.Info("Update done")
			}
		}
	}
}

func PrintMemUsage(monitor mntr.Monitor) uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mB := m.Sys / 1024 / 1024
	monitor.WithFields(map[string]interface{}{
		"MiB": mB,
	}).Info("Read current memory usage")
	return mB
}
