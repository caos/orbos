package firewall

import (
	"github.com/caos/orbiter/internal/core/operator/common"
	"github.com/caos/orbiter/internal/core/operator/nodeagent"
)

func noopEnsurer() nodeagent.FirewallEnsurer {
	return nodeagent.FirewallEnsurerFunc(func(common.Firewall) error {
		return nil
	})
}
