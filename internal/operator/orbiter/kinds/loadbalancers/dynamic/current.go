package dynamic

import (
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/internal/tree"
)

type Current struct {
	Common  *tree.Common `yaml:",inline"`
	Current struct {
		SourcePools map[string][]string
		Addresses   map[string]*infra.Address
		Desire      func(pool string, svc core.MachinesService, nodeagents map[string]*common.NodeAgentSpec, notifyMaster string) error
	} `yaml:"-"`
}
