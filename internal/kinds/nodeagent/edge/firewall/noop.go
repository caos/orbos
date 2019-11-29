package firewall

import (
	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
)

func noopEnsurer() adapter.FirewallEnsurer {
	return adapter.FirewallEnsurerFunc(func(operator.Firewall) error {
		return nil
	})
}
