package firewall

import (
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
)

func noopEnsurer() nodeagent.FirewallEnsurer {
	return nodeagent.FirewallEnsurerFunc(func(desired common.Firewall) (common.FirewallCurrent, func() error, error) {
		return nil, nil, nil
	})
}
