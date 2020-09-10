package iam

import (
	"fmt"

	zitadelbase "github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/zitadel"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
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
	err error,
) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("adapting %s failed: %w", desiredTree.Common.Kind)
		}
	}()

	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/Zitadel":
		return zitadel.AdaptFunc(timestamp, nodeselector, tolerations, features)(monitor, desiredTree, currentTree)
	default:
		return nil, nil, errors.Errorf("unknown iam kind %s", desiredTree.Common.Kind)
	}
}

func GetSecrets(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
) (
	map[string]*secret.Secret,
	error,
) {

	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/Zitadel":
		return zitadel.SecretsFunc()(monitor, desiredTree)
	default:
		return nil, errors.Errorf("unknown iam kind %s", desiredTree.Common.Kind)
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
	case "zitadel.caos.ch/Zitadel":
		return zitadel.BackupListFunc()(monitor, desiredTree)
	default:
		return nil, errors.Errorf("unknown iam kind %s", desiredTree.Common.Kind)
	}
}
