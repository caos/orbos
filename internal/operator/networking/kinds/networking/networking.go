package networking

import (
	"github.com/caos/orbos/internal/docu"
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

const (
	legacyKind = "zitadel.caos.ch/LegacyCloudflare"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	operatorLabels *labels.Operator,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	namespace string,
) (
	query core.QueryFunc,
	destroy core.DestroyFunc,
	secrets map[string]*secret.Secret,
	existing map[string]*secret.Existing,
	migrate bool,
	err error,
) {
	switch desiredTree.Common.Kind {
	case legacyKind:
		return legacycf.AdaptFunc(namespace, operatorLabels)(monitor, desiredTree, currentTree)
	default:
		return nil, nil, nil, nil, false, errors.Errorf("unknown networking kind %s", desiredTree.Common.Kind)
	}
}

func GetDocuInfo() []*docu.Type {
	infos := []*docu.Info{}

	path, versions := legacycf.GetDocuInfo()
	infos = append(infos,
		&docu.Info{
			Path:     path,
			Kind:     legacyKind,
			Versions: versions,
		},
	)

	return []*docu.Type{{
		Name:  "networking",
		Kinds: infos,
	}}
}
