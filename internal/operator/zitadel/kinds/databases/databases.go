package databases

import (
	"github.com/caos/orbos/internal/docu"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/provided"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
)

const (
	managedKind  = "zitadel.caos.ch/ManagedDatabase"
	providedKind = "zitadel.caos.ch/ProvidedDatabase"
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
	case managedKind:
		return managed.AdaptFunc(labels, users, namespace, timestamp, secretPasswordName, migrationUser, nodeselector, tolerations, features)(monitor, desiredTree, currentTree)
	case providedKind:
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
	case managedKind:
		return managed.BackupList()(monitor, desiredTree)
	case providedKind:
		return nil, errors.Errorf("no backups supported for database kind %s", desiredTree.Common.Kind)
	default:
		return nil, errors.Errorf("unknown database kind %s", desiredTree.Common.Kind)
	}
}

func GetDocuInfo() []*docu.Type {
	infos := []*docu.Info{}

	path, versions := managed.GetDocuInfo()
	infos = append(infos,
		&docu.Info{
			Path:     path,
			Kind:     managedKind,
			Versions: versions,
		},
	)

	path, versions = provided.GetDocuInfo()
	infos = append(infos,
		&docu.Info{
			Path:     path,
			Kind:     providedKind,
			Versions: versions,
		},
	)

	types := []*docu.Type{{
		Name:  "databases",
		Kinds: infos,
	}}

	types = append(types, backups.GetDocuInfo()...)
	return types
}
