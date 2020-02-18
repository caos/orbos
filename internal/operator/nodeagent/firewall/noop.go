package firewall

import (
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent"
)

func noopEnsurer() nodeagent.FirewallEnsurer {
	return nodeagent.FirewallEnsurerFunc(func(common.Firewall) (bool, error) {
		return false, nil
	})
}
