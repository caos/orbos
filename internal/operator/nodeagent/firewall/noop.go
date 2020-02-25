package firewall

import (
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent"
)

func noopEnsurer() nodeagent.FirewallEnsurer {
	return nodeagent.FirewallEnsurerFunc(func(desired common.Firewall) ([]*common.Allowed, func() error, error) {
		return nil, nil, nil
	})
}
