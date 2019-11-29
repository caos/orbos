package firewall

import (
	"github.com/caos/orbiter/internal/core/logging"
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep"
)

func Ensurer(logger logging.Logger, os dep.OperatingSystem) adapter.FirewallEnsurer {
	switch os {
	case dep.CentOS:
		return centosEnsurer(logger)
	default:
		return noopEnsurer()
	}
}
