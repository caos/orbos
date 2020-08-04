package bucket

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/secret"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups/bucket/backup"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups/bucket/clean"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups/bucket/restore"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/migration"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
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
	features []string,
) zitadel.AdaptFunc {
	return func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (queryFunc zitadel.QueryFunc, destroyFunc zitadel.DestroyFunc, err error) {
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
			checkDBReady,
		)
		queryC, destroyC, checkAndCleanupC, err := clean.ApplyFunc(
			monitor,
			name,
			namespace,
			labels,
			databases,
			checkDBReady,
		)

		queryM, destroyM, checkMigrationDone, err := migration.AdaptFunc(monitor, namespace, labels, secretPasswordName, migrationUser, users)

		destroyS, err := secret.AdaptFuncToDestroy(secretName, namespace)
		if err != nil {
			return nil, nil, err
		}

		queryS, err := secret.AdaptFuncToEnsure(secretName, namespace, labels, map[string]string{secretKey: desiredKind.Spec.ServiceAccountJSON.Value})
		if err != nil {
			return nil, nil, err
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
			nil
	}
}
