package firewall

import (
	"github.com/caos/orbiter/internal/operator/nodeagent"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep"
	"github.com/caos/orbiter/mntr"
)

func Ensurer(monitor mntr.Monitor, os dep.OperatingSystem, ignore []string) nodeagent.FirewallEnsurer {
	switch os {
	case dep.CentOS:
		return centosEnsurer(monitor, ignore)
	default:
		return noopEnsurer()
	}
}
