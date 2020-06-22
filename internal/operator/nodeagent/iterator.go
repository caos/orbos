//go:generate goderive . -dedup -autoname

package nodeagent

import (
	"fmt"
	"runtime/debug"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
)

type Rebooter interface {
	Reboot() error
}

type event struct {
	commit  string
	current *common.NodeAgentCurrent
}

func Iterator(monitor mntr.Monitor, gitClient *git.Client, nodeAgentCommit string, id string, firewallEnsurer FirewallEnsurer, conv Converter, before func() error) func() {

	doQuery := prepareQuery(monitor, nodeAgentCommit, firewallEnsurer, conv)

	return func() {
		if err := gitClient.Clone(); err != nil {
			monitor.Error(err)
			return
		}

		desired := common.NodeAgentsDesiredKind{}
		if err := yaml.Unmarshal(gitClient.Read("caos-internal/orbiter/node-agents-desired.yml"), &desired); err != nil {
			panic(err)
		}

		naDesired, ok := desired.Spec.NodeAgents.Get(id)
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

		if err := before(); err != nil {
			panic(err)
		}

		ensure, err := doQuery(*naDesired, curr)
		if err != nil {
			monitor.Error(err)
			return
		}

		readCurrent := func() common.NodeAgentsCurrentKind {
			if err := gitClient.Clone(); err != nil {
				panic(err)
			}
			current := common.NodeAgentsCurrentKind{}
			yaml.Unmarshal(gitClient.Read("caos-internal/orbiter/node-agents-current.yml"), &current)
			current.Kind = "nodeagent.caos.ch/NodeAgents"
			current.Version = "v0"
			return current
		}

		current := readCurrent()
		current.Current.Set(id, curr)

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
			current.Current.Set(id, event.current)
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

		debug.FreeOSMemory()
	}
}
