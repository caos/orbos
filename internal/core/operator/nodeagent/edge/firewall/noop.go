package firewall

import (
	"github.com/caos/orbiter/internal/core/operator/nodeagent"
	"github.com/caos/orbiter/internal/core/operator/orbiter"
)

func noopEnsurer() nodeagent.FirewallEnsurer {
	return nodeagent.FirewallEnsurerFunc(func(orbiter.Firewall) error {
		return nil
	})
}
