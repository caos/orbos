package firewall

import (
	"github.com/caos/orbos/v5/internal/operator/nodeagent"
	"github.com/caos/orbos/v5/internal/operator/nodeagent/dep"
	"github.com/caos/orbos/v5/internal/operator/nodeagent/firewall/centos"
	"github.com/caos/orbos/v5/mntr"
)

func Ensurer(monitor mntr.Monitor, os dep.OperatingSystem, open []string) nodeagent.FirewallEnsurer {
	switch os {
	case dep.CentOS:
		return centos.Ensurer(monitor, open)
	default:
		return noopEnsurer()
	}
}
