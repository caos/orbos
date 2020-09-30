package bucket

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket/backup"
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket/clean"
	"github.com/caos/orbos/internal/operator/database/kinds/backups/bucket/restore"
	"github.com/caos/orbos/mntr"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func AdaptFunc(
	name string,
	namespace string,
	labels map[string]string,
	databases []string,
	checkDBReady core.EnsureFunc,
	timestamp string,
	secretPasswordName string,
	migrationUser string,
	users []string,
	nodeselector map[string]string,
	tolerations []corev1.Toleration,
	features []string,
) core.AdaptFunc {
	return func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (queryFunc core.QueryFunc, destroyFunc core.DestroyFunc, err error) {
		secretName := "backup-serviceaccountjson"
		secretKey := "serviceaccountjson"

		internalMonitor := monitor.WithField("component", "backup")

		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
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

		//queryM, destroyM, checkMigrationDone, cleanupMigration, err := migration.AdaptFunc(monitor, namespace, "restore", labels, secretPasswordName, migrationUser, users, nodeselector, tolerations)

		destroyS, err := secret.AdaptFuncToDestroy(namespace, secretName)
		if err != nil {
			return nil, nil, err
		}

		queryS, err := secret.AdaptFuncToEnsure(namespace, secretName, labels, map[string]string{secretKey: desiredKind.Spec.ServiceAccountJSON.Value})
		if err != nil {
			return nil, nil, err
		}

		queriers := make([]core.QueryFunc, 0)
		destroyers := make([]core.DestroyFunc, 0)
		for _, feature := range features {
			switch feature {
			case "backup", "instantbackup":
				queriers = append(queriers,
					core.ResourceQueryToZitadelQuery(queryS),
					queryB,
				)
				destroyers = append(destroyers,
					core.ResourceDestroyToZitadelDestroy(destroyS),
					destroyB,
				)
			case "restore":
				queriers = append(queriers,
					queryC,
					core.EnsureFuncToQueryFunc(checkAndCleanupC),
					//queryM,
					//core.EnsureFuncToQueryFunc(checkMigrationDone),
					//core.EnsureFuncToQueryFunc(cleanupMigration),
					queryR,
					core.EnsureFuncToQueryFunc(checkAndCleanupR),
				)
				destroyers = append(destroyers,
					destroyC,
					//destroyM,
					destroyR,
				)
			}
		}

		return func(k8sClient *kubernetes2.Client, queried map[string]interface{}) (core.EnsureFunc, error) {
				return core.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
			},
			core.DestroyersToDestroyFunc(internalMonitor, destroyers),
			nil
	}
}
