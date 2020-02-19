package orbiter

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/logging/format"
)

type EnsureFunc func(psf PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error)

type commit struct {
	isErr bool
	msg   string
	files []git.File
}

func Takeoff(ctx context.Context, logger logging.Logger, gitClient *git.Client, orbiterCommit string, masterkey string, recur bool, adapt AdaptFunc) func() {

	return func() {

		treeCurrent := &Tree{}
		desiredNodeAgents := make(map[string]*common.NodeAgentSpec)

		commits := make([]*commit, 0)

		iterationLogger := logger.AddSideEffect(func(event bool, err error, fields map[string]string) {

			if !event {
				return
			}

			fields["event"] = "true"

			commits = append(commits, &commit{
				isErr: err != nil,
				msg:   format.CommitRecord(fields),
				files: []git.File{{
					Path:    "caos-internal/orbiter/current.yml",
					Content: common.MarshalYAML(treeCurrent),
				}, {
					Path: "caos-internal/orbiter/node-agents-desired.yml",
					Content: common.MarshalYAML(&common.NodeAgentsDesiredKind{
						Kind:    "nodeagent.caos.ch/NodeAgents",
						Version: "v0",
						Spec: common.NodeAgentsSpec{
							Commit:     orbiterCommit,
							NodeAgents: desiredNodeAgents,
						},
					}),
				}},
			})
		})

		treeDesired, err := parse(gitClient)
		if err != nil {
			iterationLogger.Error(err)
			return
		}

		ensure, _, _, migrate, err := adapt(iterationLogger, treeDesired, treeCurrent)
		if err != nil {
			iterationLogger.Error(err)
			return
		}

		if migrate {
			if err := pushOrbiterYML(iterationLogger, "Desired state migrated", gitClient, treeDesired); err != nil {
				logger.Error(err)
				return
			}
		}

		currentNodeAgents := common.NodeAgentsCurrentKind{}
		rawCurrentNodeAgents, _ := gitClient.Read("caos-internal/orbiter/node-agents-current.yml")
		yaml.Unmarshal(rawCurrentNodeAgents, &currentNodeAgents)

		if currentNodeAgents.Current == nil {
			currentNodeAgents.Current = make(map[string]*common.NodeAgentCurrent)
		}

		if err := ensure(pushSecretsFunc(gitClient, treeDesired), currentNodeAgents.Current, desiredNodeAgents); err != nil {
			iterationLogger.Error(err)
		}

		if err := gitClient.Clone(); err != nil {
			logger.Error(err)
		}

		for _, commit := range commits {

			changed, err := gitClient.Commit(commit.msg, commit.files...)
			if err != nil {
				panic(fmt.Errorf("Commiting event failed with err %s: %s", err.Error(), commit.msg))
			}

			if !changed && !commit.isErr {
				panic(fmt.Sprint("Event has no effect:", commit.msg))
			}
		}

		outsideChangeMessage := format.CommitRecord(map[string]string{
			"msg": "Current state changed without orbiter interaction",
		})
		changedOutside, err := gitClient.Commit(outsideChangeMessage, git.File{
			Path:    "caos-internal/orbiter/current.yml",
			Content: common.MarshalYAML(treeCurrent),
		})

		if err != nil {
			panic(fmt.Errorf("Commiting event failed with err %s: %s", err.Error(), outsideChangeMessage))
		}

		desiredOutsideMessage := format.CommitRecord(map[string]string{
			"msg": "Update desired node agents",
		})
		desiredNodeAgentsUpdated, err := gitClient.Commit(desiredOutsideMessage, git.File{
			Path: "caos-internal/orbiter/node-agents-desired.yml",
			Content: common.MarshalYAML(&common.NodeAgentsDesiredKind{
				Kind:    "nodeagent.caos.ch/NodeAgents",
				Version: "v0",
				Spec: common.NodeAgentsSpec{
					Commit:     orbiterCommit,
					NodeAgents: desiredNodeAgents,
				},
			}),
		})

		if err != nil {
			panic(fmt.Errorf("Commiting event failed with err %s: %s", err.Error(), outsideChangeMessage))
		}

		if len(commits) > 0 || changedOutside || desiredNodeAgentsUpdated {
			if err := gitClient.Push(); err != nil {
				logger.Error(err)
			}
		}
	}
}
