package loadbalancers

import (
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFunc(
	monitor mntr.Monitor,
	whitelist dynamic.WhiteListFunc,
	loadBalancingTree *tree.Tree,
	loadBalacingCurrent *tree.Tree,
	finishedChan chan bool,
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
		adaptFunc := func() (orbiter.QueryFunc, orbiter.DestroyFunc, bool, error) {
			return dynamic.AdaptFunc(whitelist)(monitor, finishedChan, loadBalancingTree, loadBalacingCurrent)
		}
		return orbiter.AdaptFuncGoroutine(adaptFunc)
	default:
		return nil, nil, false, errors.Errorf("unknown loadbalancing kind %s", loadBalancingTree.Common.Kind)
	}
}
