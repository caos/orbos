package dynamic

import (
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/v5/pkg/tree"
)

type AuthCheckResult struct {
	Machine  infra.Machine
	ExitCode int
}

type Current struct {
	Common  *tree.Common `yaml:",inline"`
	Current struct {
		Spec   func(svc core.MachinesService) (map[string][]*VIP, []AuthCheckResult, error)
		Desire func(pool string, svc core.MachinesService, vrrp *VRRP, vip func(*VIP) string) (bool, error)
	} `yaml:"-"`
}
