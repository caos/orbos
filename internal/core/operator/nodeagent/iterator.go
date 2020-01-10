//go:generate goderive . -dedup -autoname

package nodeagent

import (
	"gopkg.in/yaml.v2"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/logging"
)

type Rebooter interface {
	Reboot() error
}

func Iterator(logger logging.Logger, gitClient *git.Client, rebooter Rebooter, commit string, firewallEnsurer FirewallEnsurer, conv Converter, before func() error) func() {

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

		desired := orbiter.NodeAgentSpec{}
		if err := yaml.Unmarshal(desiredBytes, desired); err != nil {
			logger.Error(err)
			return
		}

		curr, reboot, err := ensure(logger, commit, firewallEnsurer, conv, desired)

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "internal/node-agents-current.yml", Overwrite: func(_ []byte) ([]byte, error) {
				return yaml.Marshal(curr)
			}}); err != nil {
			panic(err)
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
