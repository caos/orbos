//go:generate goderive . -dedup -autoname

package nodeagent

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/mntr"
)

type Rebooter interface {
	Reboot() error
}

type event struct {
	commit  string
	current *common.NodeAgentCurrent
}

func Iterator(monitor mntr.Monitor, gitClient *git.Client, rebooter Rebooter, nodeAgentCommit string, id string, firewallEnsurer FirewallEnsurer, conv Converter, before func() error) func() {

	return func() {
		if err := before(); err != nil {
			panic(err)
		}

		if err := gitClient.Clone(); err != nil {
			monitor.Error(err)
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
			monitor.Error(errors.New("No desired node agents found"))
			return
		}

		naDesired, ok := desired.Spec.NodeAgents[id]
		if !ok {
			monitor.Error(fmt.Errorf("No desired state for node agent with id %s found", id))
			return
		}

		if desired.Spec.Commit != nodeAgentCommit {
			monitor.WithFields(map[string]interface{}{
				"desired": desired.Spec.Commit,
				"current": nodeAgentCommit,
			}).Info("Node Agent is on the wrong commit")
			return
		}

		curr := &common.NodeAgentCurrent{}

		events := make([]*event, 0)
		monitor.OnChange = mntr.Concat(func(evt string, fields map[string]string) {
			clone := *curr
			events = append(events, &event{
				commit:  mntr.CommitRecord(mntr.AggregateCommitFields(fields)),
				current: &clone,
			})
		}, monitor.OnChange)

		ensure, err := query(monitor, nodeAgentCommit, firewallEnsurer, conv, *naDesired, curr)
		if err != nil {
			monitor.Error(err)
			return
		}

		readCurrent := func() common.NodeAgentsCurrentKind {
			if err := gitClient.Clone(); err != nil {
				panic(err)
			}

			currentNodeagents, err := gitClient.Read("caos-internal/orbiter/node-agents-current.yml")
			if err != nil {
				panic(err)
			}

			current := common.NodeAgentsCurrentKind{}
			yaml.Unmarshal(currentNodeagents, &current)
			current.Kind = "nodeagent.caos.ch/NodeAgents"
			current.Version = "v0"
			if current.Current == nil {
				current.Current = make(map[string]*common.NodeAgentCurrent)
			}
			return current
		}

		current := readCurrent()
		current.Current[id] = curr

		reconciledCurrentStateMsg := "Current state reconciled"
		reconciledCurrent, err := gitClient.StageAndCommit(mntr.CommitRecord([]*mntr.Field{{Key: "evt", Value: reconciledCurrentStateMsg}}), git.File{
			Path:    "caos-internal/orbiter/node-agents-current.yml",
			Content: common.MarshalYAML(current),
		})
		if err != nil {
			panic(fmt.Errorf("Commiting event \"%s\" failed: %s", reconciledCurrentStateMsg, err.Error()))
		}

		if reconciledCurrent {
			monitor.Error(gitClient.Push())
		}

		events = make([]*event, 0)
		if err := ensure(); err != nil {
			monitor.Error(err)
			return
		}

		current = readCurrent()

		for _, event := range events {
			current.Current[id] = event.current
			changed, err := gitClient.StageAndCommit(event.commit, git.File{
				Path:    "caos-internal/orbiter/node-agents-current.yml",
				Content: common.MarshalYAML(current),
			})
			if err != nil {
				panic(fmt.Errorf("Commiting event \"%s\" failed: %s", event.commit, err.Error()))
			}
			if !changed {
				panic(fmt.Sprint("Event has no effect:", event.commit))
			}
		}

		if len(events) > 0 {
			monitor.Error(gitClient.Push())
		}
	}
}
