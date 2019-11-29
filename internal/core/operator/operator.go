package operator

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/caos/infrop/internal/core/logging"
	"github.com/caos/infrop/internal/edge/git"
)

type Arguments struct {
	Ctx           context.Context
	Logger        logging.Logger
	MasterKey     string
	RepoURL       string
	DesiredFile   string
	CurrentFile   string
	SecretsFile   string
	DeploymentKey string
	RepoCommitter string
	RootAssembler Assembler
	Watchers      []Watcher
	Stop          chan struct{}
}

type Watcher interface {
	Watch(fire chan<- struct{}) error
}

type Iterator struct {
	args          *Arguments
	git           *git.Client
	fired         chan struct{}
	latestCurrent []byte
	secrets       map[string]interface{}
}

func New(args *Arguments) *Iterator {
	return &Iterator{
		args:  args,
		fired: make(chan struct{}),
		git:   nil,
	}
}

type IterationDone struct {
	Error   error
	Current []byte
	Secrets map[string]interface{}
}

func (i *Iterator) Initialize() error {

	i.git = git.New(i.args.Ctx, i.args.Logger, i.args.RepoCommitter, i.args.RepoURL)

	if err := i.git.Init([]byte(i.args.DeploymentKey)); err != nil {
		return errors.Wrap(err, "initializing git failed")
	}

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

	if err := i.git.Clone(); err != nil {
		return &IterationDone{Error: errors.Wrap(err, "pulling repository before iterating failed")}
	}

	current, err := i.git.Read(i.args.CurrentFile)
	if err != nil {
		return &IterationDone{Error: err}
	}

	i.secrets, err = i.git.Read(i.args.SecretsFile)
	if err != nil {
		return &IterationDone{Error: err}
	}

	rootPath, _ := i.args.RootAssembler.BuildContext()
	workDesired, workCurrent, err := toNestedRoot(i.args.Logger, i.git, rootPath, i.args.DesiredFile, current)
	if err != nil {
		return &IterationDone{Error: errors.Wrap(err, "navigating to nested root failed")}
	}

	curriedSecrets := currySecrets(i.args.Logger, func(newSecrets []byte) error {
		_, err := i.git.UpdateRemoteUntilItWorks(&git.File{
			Path: i.args.SecretsFile,
			Overwrite: func(map[string]interface{}) ([]byte, error) {
				return newSecrets, nil
			},
			Force: true,
		})
		return err
	}, i.secrets, i.args.MasterKey)

	secrets := &Secrets{curriedSecrets.read, curriedSecrets.write, curriedSecrets.delete}

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

	i.latestCurrent, err = i.git.UpdateRemoteUntilItWorks(
		&git.File{Path: i.args.CurrentFile, Overwrite: func(reloadedCurrent map[string]interface{}) ([]byte, error) {
			_, reloadedWorkCurrent, err := toNestedRoot(i.args.Logger, i.git, rootPath, i.args.DesiredFile, reloadedCurrent)
			if err != nil {
				return nil, errors.Wrap(err, "navigating to reloaded nested root failed")
			}

			if err := rebuildCurrent(i.args.Logger, reloadedWorkCurrent, tree); err != nil {
				return nil, errors.Wrap(err, "overwriting current state failed")
			}

			return marshal(reloadedCurrent)
		}})
	if err != nil {
		return &IterationDone{Error: err}
	}

	return &IterationDone{Current: i.latestCurrent, Secrets: i.secrets}
}
