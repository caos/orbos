package databases

import (
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/provided"
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
	namespace string,
	users []string,
	labels map[string]string,
	timestamp string,
	secretPasswordName string,
	migrationUser string,
	nodeselector map[string]string,
	tolerations []core.Toleration,
	features []string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	map[string]*secret.Secret,
	error,
) {
	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/ManagedDatabase":
		return managed.AdaptFunc(labels, users, namespace, timestamp, secretPasswordName, migrationUser, nodeselector, tolerations, features)(monitor, desiredTree, currentTree)
	case "zitadel.caos.ch/ProvidedDatabse":
		return provided.AdaptFunc()(monitor, desiredTree, currentTree)
	default:
		return nil, nil, nil, errors.Errorf("unknown database kind %s", desiredTree.Common.Kind)
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
