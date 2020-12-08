package databases

import (
	core2 "github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/provided"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	namespace string,
	operatorLabels *labels.Operator,
	timestamp string,
	nodeselector map[string]string,
	tolerations []core.Toleration,
	version string,
	features []string,
) (
	core2.QueryFunc,
	core2.DestroyFunc,
	map[string]*secret.Secret,
	error,
) {
	switch desiredTree.Common.Kind {
	case "databases.caos.ch/CockroachDB":
		return managed.AdaptFunc(operatorLabels, labels.MustForAPI(operatorLabels, "CockroachDB", desiredTree.Common.Version), namespace, timestamp, nodeselector, tolerations, version, features)(monitor, desiredTree, currentTree)
	case "databases.caos.ch/ProvidedDatabse":
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
	case "databases.caos.ch/CockroachDB":
		return managed.BackupList()(monitor, desiredTree)
	case "databases.caos.ch/ProvidedDatabse":
		return nil, errors.Errorf("no backups supported for database kind %s", desiredTree.Common.Kind)
	default:
		return nil, errors.Errorf("unknown database kind %s", desiredTree.Common.Kind)
	}
}
