package loadbalancers

import (
	"github.com/caos/orbos/internal/docu"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

const (
	dynamicKind = "orbiter.caos.ch/DynamicLoadBalancer"
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
	case dynamicKind:
		return dynamic.AdaptFunc(whitelist)(monitor, finishedChan, loadBalancingTree, loadBalacingCurrent)
	default:
		return nil, nil, nil, false, nil, errors.Errorf("unknown loadbalancing kind %s", loadBalancingTree.Common.Kind)
	}
}

func GetDocuInfo() []*docu.Type {
	path, dynVersions := dynamic.GetDocuInfo()
	return []*docu.Type{{
		Name: "loadbalancing",
		Kinds: []*docu.Info{
			{
				Path:     path,
				Kind:     dynamicKind,
				Versions: dynVersions,
			},
		},
	}}
}
