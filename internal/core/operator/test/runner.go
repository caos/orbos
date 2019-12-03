package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/edge/git"
	logcontext "github.com/caos/orbiter/logging/context"
	"github.com/caos/orbiter/logging/stdlib"
	"github.com/caos/orbiter/watcher/cron"
	"github.com/caos/orbiter/watcher/immediate"
	"github.com/caos/orbiter/internal/kinds/orbiter"
	"github.com/caos/orbiter/internal/kinds/orbiter/adapter"
	"github.com/caos/orbiter/internal/kinds/orbiter/model"
)

const testRepoURL = "git@github.com:caos/orbiter-test.git"

type Desire func(map[string]interface{})

func (d Desire) Chain(prioritized Desire) Desire {
	return func(des map[string]interface{}) {
		d(des)
		prioritized(des)
	}
}

type IterationResult struct {
	Error   error
	Current *Current
	Secrets map[string]interface{}
}

type Current struct {
	Deps map[string]struct {
		Current struct {
			Kind  string
			State struct {
				Status   string
				Computes map[string]struct {
					Status   string
					Metadata struct {
						Tier string
					}
					Software struct {
						Current struct {
							State struct {
								Version  string
								Ready    bool
								Software struct {
									Kubelet struct {
										Version string
										Config  map[string]string
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func Run(
	stop chan struct{},
	phase string,
	t *testing.T,
	desiredFile string,
	desire func(map[string]interface{})) (iterations chan *IterationResult, cleanup func() error, logger logging.Logger, err error) {

	rootContext := context.Background()
	runContext, cancelRunning := context.WithCancel(rootContext)

	go func() {
		<-stop
		cancelRunning()
	}()

	loggerFields := map[string]interface{}{
		"phase": phase,
	}
	logger = logcontext.Add(stdlib.New(os.Stderr)).WithFields(loggerFields)

	cleanup = func() error { return nil }

	gitClient := git.New(runContext, logger, t.Name(), testRepoURL)

	masterKey, err := ioutil.ReadFile("/etc/orbiter/masterkey")
	if err != nil {
		return nil, cleanup, logger, err
	}

	repoKey, err := ioutil.ReadFile("/etc/orbiter/repokey")
	if err != nil {
		return nil, cleanup, logger, err
	}

	if err = gitClient.Init([]byte(repoKey)); err != nil {
		return nil, cleanup, logger, err
	}

	if err := gitClient.Clone(); err != nil {
		return nil, cleanup, logger, err
	}

	desiredYml := "desired.yml"
	currentYml := "current.yml"
	secretsYml := "secrets.yml"

	if _, err := gitClient.UpdateRemoteUntilItWorks(&git.File{
		Path: desiredYml,
		Overwrite: func(_ map[string]interface{}) ([]byte, error) {
			return []byte(desiredFile), nil
		},
		Force: true,
	}); err != nil {
		return nil, cleanup, logger, err
	}

	testOp := operator.New(&operator.Arguments{
		Ctx:           runContext,
		Logger:        logger,
		MasterKey:     string(masterKey),
		RepoURL:       testRepoURL,
		DesiredFile:   desiredYml,
		CurrentFile:   currentYml,
		SecretsFile:   secretsYml,
		DeploymentKey: string(repoKey),
		RepoCommitter: "Orbiter",
		Watchers: []operator.Watcher{
			immediate.New(logger),
			cron.New(logger, "@every 30s"),
		},
		RootAssembler: orbiter.New(nil, desire, adapter.New(&model.Config{
			Logger:           logger,
			ConfigID:         "orbitertest",
			NodeagentRepoURL: testRepoURL,
			NodeagentRepoKey: string(repoKey),
			OrbiterVersion:    "dev",
			CurrentFile:      currentYml,
			SecretsFile:      secretsYml,
			Masterkey:        string(masterKey),
		})),
	})

	if err := testOp.Initialize(); err != nil {
		return nil, cleanup, logger, err
	}

	marshalledIterations := make(chan *operator.IterationDone)

	go testOp.Run(marshalledIterations)

	var (
		latestCurrent []byte
		latestSecrets map[string]interface{}
	)

	var mux sync.Mutex

	iterations = make(chan *IterationResult)
	go func() {

		for it := range marshalledIterations {
			mux.Lock()
			latestCurrent = it.Current
			latestSecrets = it.Secrets
			mux.Unlock()
			iterations <- MapIteration(it)
		}
		close(iterations)
	}()

	return iterations, func() error {

		cancelRunning()

		if t.Failed() {
			t.Log("Test failed. Printing latest known secrets and current state before starting the cleanup process")
			mux.Lock()
			t.Log(fmt.Sprintf("%#+v", latestSecrets))
			t.Log(string(latestCurrent))
			mux.Unlock()
		}

		cleanupCtx, cancelCleanupping := context.WithTimeout(rootContext, 5*time.Minute)
		defer cancelCleanupping()
		stopCleanup := make(chan struct{})
		cleanupIteration, _, _, err := Run(stopCleanup, phase+" cleanup", t, desiredFile, func(des map[string]interface{}) {
			des["destroyed"] = true
		})

		if err != nil {
			return err
		}
		select {
		case it := <-cleanupIteration:
			return it.Error
		case <-cleanupCtx.Done():
			return cleanupCtx.Err()
		}
	}, logger, nil
}

func MapIteration(it *operator.IterationDone) *IterationResult {

	current := &Current{}
	if err := yaml.Unmarshal(it.Current, current); err != nil {
		panic(err)
	}
	return &IterationResult{
		Error:   it.Error,
		Current: current,
		Secrets: it.Secrets,
	}
}
