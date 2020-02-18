//go:generate goderive . -dedup -autoname

package nodeagent

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/logging/format"
)

type Rebooter interface {
	Reboot() error
}

type commit struct {
	msg     string
	current common.NodeAgentCurrent
}

func Iterator(logger logging.Logger, gitClient *git.Client, rebooter Rebooter, nodeAgentCommit string, id string, firewallEnsurer FirewallEnsurer, conv Converter, before func() error) func() {

	return func() {
		if err := before(); err != nil {
			panic(err)
		}

		if err := gitClient.Clone(); err != nil {
			logger.Error(err)
			return
		}

		desiredBytes, err := gitClient.Read("caos-internal/orbiter/node-agents-desired.yml")
		if err != nil {
			panic(err)
		}

		desired := common.NodeAgentsDesiredKind{}
		if err := yaml.Unmarshal(desiredBytes, &desired); err != nil {
			panic(err)
		}

		if desired.Spec.NodeAgents == nil {
			logger.Error(errors.New("No desired node agents found"))
			return
		}

		naDesired, ok := desired.Spec.NodeAgents[id]
		if !ok {
			logger.Error(fmt.Errorf("No desired state for node agent with id %s found", id))
			return
		}

		if desired.Spec.Commit != nodeAgentCommit {
			logger.WithFields(map[string]interface{}{
				"desired": desired.Spec.Commit,
				"current": nodeAgentCommit,
			}).Info(false, "Node Agent is on the wrong commit")
			return
		}

		curr := &common.NodeAgentCurrent{}
		commits := make([]*commit, 0)
		if err := ensure(logger.AddSideEffect(func(event bool, fields map[string]string) {

			if !event {
				return
			}

			fields["event"] = "true"

			snap := *curr
			commits = append(commits, &commit{
				msg:     format.CommitRecord(fields),
				current: snap,
			})
		}), nodeAgentCommit, firewallEnsurer, conv, *naDesired, curr); err != nil {
			logger.Error(err)
			return
		}

		if err := gitClient.Clone(); err != nil {
			logger.Error(err)
			return
		}

		currentNodeagents, err := gitClient.Read("caos-internal/orbiter/node-agents-current.yml")
		if err != nil {
			panic(err)
		}

		doCommit := func(commit commit) bool {
			current := common.NodeAgentsCurrentKind{}
			if err := yaml.Unmarshal(currentNodeagents, &current); err != nil {
				panic(err)
			}
			current.Kind = "nodeagent.caos.ch/NodeAgents"
			current.Version = "v0"
			if current.Current == nil {
				current.Current = make(map[string]*common.NodeAgentCurrent)
			}
			current.Current[id] = &commit.current

			changed, err := gitClient.Commit(commit.msg, git.File{
				Path:    "caos-internal/orbiter/node-agents-current.yml",
				Content: common.MarshalYAML(current),
			})

			if err != nil {
				panic(fmt.Errorf("Commiting event failed with err %s: %s", err.Error(), commit.msg))
			}
			return changed
		}

		for _, commit := range commits {
			if !doCommit(*commit) {
				panic(fmt.Sprint("Event has no effect:", commit.msg))
			}
		}

		if len(commits) > 0 || doCommit(commit{
			msg:     "Current state changed without node agent interaction",
			current: *curr,
		}) {
			if err := gitClient.Push(); err != nil {
				logger.Error(err)
			}
		}
	}
}
