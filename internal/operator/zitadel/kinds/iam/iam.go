package iam

import (
	"fmt"
	"github.com/caos/orbos/internal/docu"
	zitadelbase "github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
)

const (
	zitadelKind = "zitadel.caos.ch/Zitadel"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	nodeselector map[string]string,
	tolerations []core.Toleration,
	timestamp string,
	features ...string,
) (
	query zitadelbase.QueryFunc,
	destroy zitadelbase.DestroyFunc,
	secrets map[string]*secret.Secret,
	err error,
) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("adapting %s failed: %w", desiredTree.Common.Kind)
		}
	}()

	switch desiredTree.Common.Kind {
	case zitadelKind:
		return zitadel.AdaptFunc(timestamp, nodeselector, tolerations, features)(monitor, desiredTree, currentTree)
	default:
		return nil, nil, nil, errors.Errorf("unknown iam kind %s", desiredTree.Common.Kind)
	}
}

func GetBackupList(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
) (
	[]string,
	error,
) {
	switch desiredTree.Common.Kind {
	case zitadelKind:
		return zitadel.BackupListFunc()(monitor, desiredTree)
	default:
		return nil, errors.Errorf("unknown iam kind %s", desiredTree.Common.Kind)
	}
}

func GetDocuInfo() []*docu.Type {
	path, versions := zitadel.GetDocuInfo()
	types := []*docu.Type{{
		Name: "iam",
		Kinds: []*docu.Info{{
			Path:     path,
			Kind:     zitadelKind,
			Versions: versions,
		}},
	}}

	types = append(types, databases.GetDocuInfo()...)
	types = append(types, networking.GetDocuInfo()...)
	return types
}
