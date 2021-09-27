package networking

import (
	"github.com/caos/orbos/v5/internal/operator/nodeagent"
	"github.com/caos/orbos/v5/internal/operator/nodeagent/dep"
	"github.com/caos/orbos/v5/internal/operator/nodeagent/networking/centos"
	"github.com/caos/orbos/v5/mntr"
)

func Ensurer(monitor mntr.Monitor, os dep.OperatingSystem) nodeagent.NetworkingEnsurer {
	switch os {
	case dep.CentOS:
		return centos.Ensurer(monitor)
	default:
		return noopEnsurer()
	}
}
