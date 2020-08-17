package databases

import (
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/provided"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	caCertificate string,
	caKey string,
	dbs []string,
	namespace string,
	users []string,
	labels map[string]string,
	timestamp string,
	secretPasswordName string,
	migrationUser string,
	features []string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/ManagedDatabase":
		return managed.AdaptFunc(caCertificate, caKey, dbs, labels, users, namespace, timestamp, secretPasswordName, migrationUser, features)(monitor, desiredTree, currentTree)
	case "zitadel.caos.ch/ProvidedDatabase":
		return provided.AdaptFunc()(monitor, desiredTree, currentTree)
	default:
		return nil, nil, errors.Errorf("unknown database kind %s", desiredTree.Common.Kind)
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
