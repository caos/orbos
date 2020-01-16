//go:generate goderive . -dedup -autoname

package nodeagent

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/core/operator/common"
	"github.com/caos/orbiter/internal/edge/git"
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

		desiredBytes, err := gitClient.Read("internal/node-agents-desired.yml")
		if err != nil {
			logger.Error(err)
			return
		}

		desired := common.NodeAgentsSpec{}
		if err := yaml.Unmarshal(desiredBytes, desired); err != nil {
			logger.Error(err)
			return
		}

		if desired.NodeAgents == nil {
			logger.Error(errors.New("No desired node agents found"))
			return
		}

		naDesired, ok := desired.NodeAgents[id]
		if !ok {
			logger.Error(fmt.Errorf("No desired state for node agent with id %s found", id))
			return
		}

		curr, reboot, err := ensure(logger, commit, firewallEnsurer, conv, *naDesired)
		if err != nil {
			logger.Error(err)
			return
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "internal/node-agents-current.yml", Overwrite: func(nodeagents []byte) ([]byte, error) {
				current := common.NodeAgentsCurrentKind{}
				if err := yaml.Unmarshal(nodeagents, current); err != nil {
					return nil, err
				}
				if current.Current == nil {
					current.Current = make(map[string]*common.NodeAgentCurrent)
				}
				current.Current[id] = curr

				return common.MarshalYAML(current), nil
			}}); err != nil {
			logger.Error(err)
			return
		}

		if reboot {
			logger.Info("Rebooting")
			if err = rebooter.Reboot(); err != nil {
				logger.Error(err)
				return
			}
		}
	}
}
