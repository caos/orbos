package operator

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/logging"
)

type Arguments struct {
	Ctx             context.Context
	Logger          logging.Logger
	GitClient       *git.Client
	MasterKey       string
	DesiredFile     string
	CurrentFile     string
	SecretsFile     string
	RootAssembler   Assembler
	Watchers        []Watcher
	Stop            chan struct{}
	BeforeIteration func(desired []byte, secrets *Secrets) error
}

type Watcher interface {
	Watch(fire chan<- struct{}) error
}

type Iterator struct {
	args          *Arguments
	fired         chan struct{}
	latestCurrent []byte
	secrets       map[string]interface{}
}

func New(args *Arguments) *Iterator {
	return &Iterator{
		args:    args,
		fired:   make(chan struct{}),
		secrets: make(map[string]interface{}),
	}
}

type IterationDone struct {
	Error   error
	Current []byte
	Secrets map[string]interface{}
}

func (i *Iterator) Initialize() error {

	for _, watcher := range i.args.Watchers {
		if err := watcher.Watch(i.fired); err != nil {
			return errors.Wrap(err, "starting watcher failed")
		}
	}
	return nil
}

func (i *Iterator) Run(iterations chan<- *IterationDone) {
	//	refspecs := []config.RefSpec{config.RefSpec("+refs/heads/master:refs/remotes/origin/master")}

loop:
	for {
		select {
		case <-i.args.Ctx.Done():
			return
		case <-i.fired:
			if len(i.fired) != 0 {
				i.args.Logger.Info("Skipping iteration")
				continue loop
			}
			iterations <- i.iterate(i.args.Ctx.Done())
		}
	}
}

func (i *Iterator) iterate(stop <-chan struct{}) *IterationDone {

	started := time.Now()

	defer func() {
		i.args.Logger.WithFields(map[string]interface{}{
			"took": time.Now().Sub(started),
		}).Info("Iteration done")
	}()

	if err := i.args.GitClient.Clone(); err != nil {
		return &IterationDone{Error: errors.Wrap(err, "pulling repository before iterating failed")}
	}

	desiredBytes, err := i.args.GitClient.Read(i.args.DesiredFile)
	if err != nil {
		return &IterationDone{Error: err}
	}

	secretsBytes, err := i.args.GitClient.Read(i.args.SecretsFile)
	if err != nil {
		return &IterationDone{Error: err}
	}

	if err := yaml.Unmarshal(secretsBytes, &i.secrets); err != nil {
		return &IterationDone{Error: err}
	}

	curriedSecrets := currySecrets(i.args.Logger, func(newSecrets map[string]interface{}) error {
		_, err := i.args.GitClient.UpdateRemoteUntilItWorks(&git.File{
			Path: i.args.SecretsFile,
			Overwrite: func([]byte) ([]byte, error) {
				return Marshal(newSecrets)
			},
			Force: true,
		})
		return err
	}, i.secrets, i.args.MasterKey)

	secrets := &Secrets{curriedSecrets.read, curriedSecrets.write, curriedSecrets.delete}

	if i.args.BeforeIteration != nil {
		if err := i.args.BeforeIteration(desiredBytes, secrets); err != nil {
			return &IterationDone{Error: err}
		}
	}

	currentBytes, err := i.args.GitClient.Read(i.args.CurrentFile)
	if err != nil {
		return &IterationDone{Error: err}
	}

	desired := make(map[string]interface{})
	current := make(map[string]interface{})

	if err := yaml.Unmarshal(desiredBytes, &desired); err != nil {
		return &IterationDone{Error: err}
	}

	if err := yaml.Unmarshal(currentBytes, &current); err != nil {
		return &IterationDone{Error: err}
	}

	rootPath, _ := i.args.RootAssembler.BuildContext()
	workDesired, workCurrent, err := toNestedRoot(i.args.Logger, i.args.GitClient, rootPath, desired, current)
	if err != nil {
		return &IterationDone{Error: errors.Wrap(err, "navigating to nested root failed")}
	}

	select {
	case <-i.args.Ctx.Done():
		return &IterationDone{Current: i.latestCurrent, Secrets: i.secrets}
	default:
		// continue
	}

	tree, err := build(i.args.Logger, i.args.RootAssembler, workDesired, workCurrent, secrets, nil, true)
	if err != nil {
		return &IterationDone{Error: err}
	}

	if _, err := ensure(i.args.Ctx, i.args.Logger, tree, secrets); err != nil {
		return &IterationDone{Error: err}
	}

	i.latestCurrent, err = i.args.GitClient.UpdateRemoteUntilItWorks(
		&git.File{Path: i.args.CurrentFile, Overwrite: func(reloadedCurrent []byte) ([]byte, error) {

			reloadedCurrentMap := make(map[string]interface{})
			if err := yaml.Unmarshal(reloadedCurrent, &reloadedCurrentMap); err != nil {
				return nil, err
			}

			_, reloadedWorkCurrent, err := toNestedRoot(i.args.Logger, i.args.GitClient, rootPath, desired, reloadedCurrentMap)
			if err != nil {
				return nil, errors.Wrap(err, "navigating to reloaded nested root failed")
			}

			if err := rebuildCurrent(i.args.Logger, reloadedWorkCurrent, tree); err != nil {
				return nil, errors.Wrap(err, "overwriting current state failed")
			}

			return Marshal(reloadedCurrentMap)
		}})
	if err != nil {
		return &IterationDone{Error: err}
	}

	return &IterationDone{Current: i.latestCurrent, Secrets: i.secrets}
}
