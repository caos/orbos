//go:generate goderive . -dedup -autoname

package nodeagent

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/logging"
)

type Rebooter interface {
	Reboot() error
}

func Iterator(logger logging.Logger, gitClient *git.Client, rebooter Rebooter, commit string, id string, firewallEnsurer FirewallEnsurer, conv Converter, before func() error) func() {

	return func() {
		if err := before(); err != nil {
			panic(err)
		}

		if err := gitClient.Clone(); err != nil {
			panic(err)
		}

		desiredBytes, err := gitClient.Read("caos-internal/orbiter/node-agents-desired.yml")
		if err != nil {
			logger.Error(err)
			return
		}

		desired := common.NodeAgentsDesiredKind{}
		if err := yaml.Unmarshal(desiredBytes, &desired); err != nil {
			logger.Error(err)
			return
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

		if desired.Spec.Commit != commit {
			logger.WithFields(map[string]interface{}{
				"desired": desired.Spec.Commit,
				"current": commit,
			}).Info("Node Agent is on the wrong commit")
			return
		}

		curr, err := ensure(logger, commit, firewallEnsurer, conv, *naDesired)
		if err != nil {
			logger.Error(err)
			return
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks("caos-internal/orbiter/node-agents-current.yml", func(nodeagents []byte) ([]byte, error) {
			current := common.NodeAgentsCurrentKind{}
			if err := yaml.Unmarshal(nodeagents, &current); err != nil {
				return nil, err
			}
			current.Kind = "nodeagent.caos.ch/NodeAgents"
			current.Version = "v0"
			if current.Current == nil {
				current.Current = make(map[string]*common.NodeAgentCurrent)
			}
			current.Current[id] = curr

			return common.MarshalYAML(current), nil
		}, true); err != nil {
			logger.Error(err)
			return
		}
	}
}
