package dynamic

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/pkg/tree"
)

type Current struct {
	Common  *tree.Common `yaml:",inline"`
	Current struct {
		Spec   func(svc core.MachinesService) (map[string][]*VIP, error)
		Desire func(pool string, svc core.MachinesService, vrrp bool, notifyMaster func(machine infra.Machine, peers infra.Machines, vips []*VIP) string, vip func(*VIP) string) (bool, error)
	} `yaml:"-"`
}
