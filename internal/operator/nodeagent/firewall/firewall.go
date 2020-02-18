package firewall

import (
	"github.com/caos/orbiter/internal/operator/nodeagent"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep"
	"github.com/caos/orbiter/logging"
)

func Ensurer(logger logging.Logger, os dep.OperatingSystem, ignore []string) nodeagent.FirewallEnsurer {
	switch os {
	case dep.CentOS:
		return centosEnsurer(logger, ignore)
	default:
		return noopEnsurer()
	}
}
