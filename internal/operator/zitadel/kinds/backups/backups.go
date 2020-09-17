package backups

import (
	"github.com/caos/orbos/internal/docu"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups/bucket"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	bucketKind = "zitadel.caos.ch/BucketBackup"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	name string,
	namespace string,
	labels map[string]string,
	databases []string,
	checkDBReady zitadel.EnsureFunc,
	timestamp string,
	secretPasswordName string,
	migrationUser string,
	users []string,
	nodeselector map[string]string,
	tolerations []corev1.Toleration,
	features []string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	map[string]*secret.Secret,
	error,
) {
	switch desiredTree.Common.Kind {
	case bucketKind:
		return bucket.AdaptFunc(name, namespace, labels, databases, checkDBReady, timestamp, secretPasswordName, migrationUser, users, nodeselector, tolerations, features)(monitor, desiredTree, currentTree)
	default:
		return nil, nil, nil, errors.Errorf("unknown database kind %s", desiredTree.Common.Kind)
	}
}

func GetBackupList(
	monitor mntr.Monitor,
	name string,
	desiredTree *tree.Tree,
) (
	[]string,
	error,
) {
	switch desiredTree.Common.Kind {
	case bucketKind:
		return bucket.BackupList()(monitor, name, desiredTree)
	default:
		return nil, errors.Errorf("unknown database kind %s", desiredTree.Common.Kind)
	}
}

func GetDocuInfo() []*docu.Type {
	path, versions := bucket.GetDocuInfo()
	return []*docu.Type{{
		Name: "backup",
		Kinds: []*docu.Info{
			{
				Path:     path,
				Kind:     bucketKind,
				Versions: versions,
			},
		},
	}}
}
