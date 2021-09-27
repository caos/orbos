package loadbalancers

import (
	"fmt"

	"github.com/caos/orbos/v5/internal/operator/orbiter"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/secret"
	"github.com/caos/orbos/v5/pkg/tree"
)

func GetQueryAndDestroyFunc(
	monitor mntr.Monitor,
	whitelist dynamic.WhiteListFunc,
	loadBalancingTree *tree.Tree,
	loadBalacingCurrent *tree.Tree,
	finishedChan chan struct{},
) (
	orbiter.QueryFunc,
	orbiter.DestroyFunc,
	orbiter.ConfigureFunc,
	bool,
	map[string]*secret.Secret,
	error,
) {

	switch loadBalancingTree.Common.Kind {
	//		case "orbiter.caos.ch/ExternalLoadBalancer":
	//			return []orbiter.Assembler{external.New(depPath, generalOverwriteSpec, externallbadapter.New())}, nil
	case "orbiter.caos.ch/DynamicLoadBalancer":
		return dynamic.AdaptFunc(whitelist)(monitor, finishedChan, loadBalancingTree, loadBalacingCurrent)
	default:
		return nil, nil, nil, false, nil, mntr.ToUserError(fmt.Errorf("unknown loadbalancing kind %s", loadBalancingTree.Common.Kind))
	}
}
