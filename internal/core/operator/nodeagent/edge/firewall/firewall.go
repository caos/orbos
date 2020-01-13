package firewall

import (
	"github.com/caos/orbiter/internal/core/operator/nodeagent"
	"github.com/caos/orbiter/internal/core/operator/nodeagent/edge/dep"
	"github.com/caos/orbiter/logging"
)

func Ensurer(logger logging.Logger, os dep.OperatingSystem) nodeagent.FirewallEnsurer {
	switch os {
	case dep.CentOS:
		return centosEnsurer(logger)
	default:
		return noopEnsurer()
	}
}
