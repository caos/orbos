package loadbalancers

import (
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/tree"
	"github.com/caos/orbiter/mntr"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFunc(
	monitor mntr.Monitor,
	whitelist dynamic.WhiteListFunc,
	loadBalancingTree *tree.Tree,
	loadBalacingCurrent *tree.Tree,
) (
	orbiter.QueryFunc,
	orbiter.DestroyFunc,
	bool,
	error,
) {

	switch loadBalancingTree.Common.Kind {
	//		case "orbiter.caos.ch/ExternalLoadBalancer":
	//			return []orbiter.Assembler{external.New(depPath, generalOverwriteSpec, externallbadapter.New())}, nil
	case "orbiter.caos.ch/DynamicLoadBalancer":
		return dynamic.AdaptFunc(whitelist)(monitor, loadBalancingTree, loadBalacingCurrent)
	default:
		return nil, nil, false, errors.Errorf("unknown loadbalancing kind %s", loadBalancingTree.Common.Kind)
	}
}
