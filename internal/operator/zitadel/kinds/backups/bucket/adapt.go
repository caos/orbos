package bucket

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/secret"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups/bucket/backup"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups/bucket/clean"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups/bucket/restore"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/zitadel/migration"
	orbossecret "github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func AdaptFunc(
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
) zitadel.AdaptFunc {

	secretName := "backup-serviceaccountjson"
	secretKey := "serviceaccountjson"

	return func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (queryFunc zitadel.QueryFunc, destroyFunc zitadel.DestroyFunc, secrets map[string]*orbossecret.Secret, err error) {

		internalMonitor := monitor.WithField("component", "backup")

		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		if !monitor.IsVerbose() && desiredKind.Spec.Verbose {
			internalMonitor.Verbose()
		}

		queryB, destroyB, err := backup.AdaptFunc(
			internalMonitor,
			name,
			namespace,
			labels,
			databases,
			checkDBReady,
			desiredKind.Spec.Bucket,
			desiredKind.Spec.Cron,
			secretName,
			secretKey,
			timestamp,
			nodeselector,
			tolerations,
			features,
		)

		queryR, destroyR, checkAndCleanupR, err := restore.ApplyFunc(
			monitor,
			name,
			namespace,
			labels,
			databases,
			desiredKind.Spec.Bucket,
			timestamp,
			nodeselector,
			tolerations,
			checkDBReady,
		)

		queryC, destroyC, checkAndCleanupC, err := clean.ApplyFunc(
			monitor,
			name,
			namespace,
			labels,
			databases,
			nodeselector,
			tolerations,
			checkDBReady,
		)

		queryM, destroyM, checkMigrationDone, cleanupMigration, err := migration.AdaptFunc(monitor, namespace, "restore", labels, secretPasswordName, migrationUser, users, nodeselector, tolerations)

		destroyS, err := secret.AdaptFuncToDestroy(namespace, secretName)
		if err != nil {
			return nil, nil, nil, err
		}

		queryS, err := secret.AdaptFuncToEnsure(namespace, secretName, labels, map[string]string{secretKey: desiredKind.Spec.ServiceAccountJSON.Value})
		if err != nil {
			return nil, nil, nil, err
		}

		queriers := make([]zitadel.QueryFunc, 0)
		destroyers := make([]zitadel.DestroyFunc, 0)
		for _, feature := range features {
			switch feature {
			case "backup", "instantbackup":
				queriers = append(queriers,
					zitadel.ResourceQueryToZitadelQuery(queryS),
					queryB,
				)
				destroyers = append(destroyers,
					zitadel.ResourceDestroyToZitadelDestroy(destroyS),
					destroyB,
				)
			case "restore":
				queriers = append(queriers,
					queryC,
					zitadel.EnsureFuncToQueryFunc(checkAndCleanupC),
					queryM,
					zitadel.EnsureFuncToQueryFunc(checkMigrationDone),
					zitadel.EnsureFuncToQueryFunc(cleanupMigration),
					queryR,
					zitadel.EnsureFuncToQueryFunc(checkAndCleanupR),
				)
				destroyers = append(destroyers,
					destroyC,
					destroyM,
					destroyR,
				)
			}
		}

		return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
				return zitadel.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
			},
			zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
			getSecretsMap(desiredKind),
			nil
	}
}
