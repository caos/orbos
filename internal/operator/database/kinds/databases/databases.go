package databases

import (
	core2 "github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/provided"
	"github.com/caos/orbos/mntr"
	secret2 "github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	namespace string,
	labels map[string]string,
	timestamp string,
	nodeselector map[string]string,
	tolerations []core.Toleration,
	features []string,
) (
	core2.QueryFunc,
	core2.DestroyFunc,
	error,
) {
	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/ManagedDatabase":
		return managed.AdaptFunc(labels, namespace, timestamp, nodeselector, tolerations, features)(monitor, desiredTree, currentTree)
	case "zitadel.caos.ch/ProvidedDatabse":
		return provided.AdaptFunc()(monitor, desiredTree, currentTree)
	default:
		return nil, nil, errors.Errorf("unknown database kind %s", desiredTree.Common.Kind)
	}
}

func GetSecrets(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
) (
	map[string]*secret2.Secret,
	error,
) {

	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/ManagedDatabase":
		return managed.SecretsFunc()(monitor, desiredTree)
	case "zitadel.caos.ch/ProvidedDatabse":
		return provided.SecretsFunc()(monitor, desiredTree)
	default:
		return nil, errors.Errorf("unknown database kind %s", desiredTree.Common.Kind)
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
	case "zitadel.caos.ch/ManagedDatabase":
		return managed.BackupList()(monitor, desiredTree)
	case "zitadel.caos.ch/ProvidedDatabse":
		return nil, errors.Errorf("no backups supported for database kind %s", desiredTree.Common.Kind)
	default:
		return nil, errors.Errorf("unknown database kind %s", desiredTree.Common.Kind)
	}
}
