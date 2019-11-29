package firewall

import (
	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/nodeagent/adapter"
)

func noopEnsurer() adapter.FirewallEnsurer {
	return adapter.FirewallEnsurerFunc(func(operator.Firewall) error {
		return nil
	})
}
