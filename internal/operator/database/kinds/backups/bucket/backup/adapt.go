package backup

import (
	"time"

	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/cronjob"
	"github.com/caos/orbos/pkg/kubernetes/resources/job"
	"github.com/caos/orbos/pkg/labels"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultMode        int32         = 256
	certPath                         = "/cockroach/cockroach-certs"
	secretPath                       = "/secrets/sa.json"
	backupPath                       = "/cockroach"
	backupNameEnv                    = "BACKUP_NAME"
	cronJobNamePrefix                = "backup-"
	internalSecretName               = "client-certs"
	image                            = "ghcr.io/caos/crbackup"
	rootSecretName                   = "cockroachdb.client.root"
	timeout            time.Duration = 60
	Normal                           = "backup"
	Instant                          = "instantbackup"
)

func AdaptFunc(
	monitor mntr.Monitor,
	backupName string,
	namespace string,
	componentLabels *labels.Component,
	databases []string,
	checkDBReady core.EnsureFunc,
	bucketName string,
	cron string,
	secretName string,
	secretKey string,
	timestamp string,
	nodeselector map[string]string,
	tolerations []corev1.Toleration,
	features []string,
	version string,
) (
	queryFunc core.QueryFunc,
	destroyFunc core.DestroyFunc,
	err error,
) {

	command := getBackupCommand(
		timestamp,
		databases,
		bucketName,
		backupName,
	)

	jobSpecDef := getJobSpecDef(
		nodeselector,
		tolerations,
		secretName,
		secretKey,
		backupName,
		version,
		command,
	)

	destroyers := []core.DestroyFunc{}
	queriers := []core.QueryFunc{}

	cronJobDef := getCronJob(
		namespace,
		labels.MustForName(componentLabels, GetJobName(backupName)),
		cron,
		jobSpecDef,
	)

	destroyCJ, err := cronjob.AdaptFuncToDestroy(cronJobDef.Namespace, cronJobDef.Name)
	if err != nil {
		return nil, nil, err
	}

	queryCJ, err := cronjob.AdaptFuncToEnsure(cronJobDef)
	if err != nil {
		return nil, nil, err
	}

	jobDef := getJob(
		namespace,
		labels.MustForName(componentLabels, cronJobNamePrefix+backupName),
		jobSpecDef,
	)

	destroyJ, err := job.AdaptFuncToDestroy(jobDef.Namespace, jobDef.Name)
	if err != nil {
		return nil, nil, err
	}

	queryJ, err := job.AdaptFuncToEnsure(jobDef)
	if err != nil {
		return nil, nil, err
	}

	for _, feature := range features {
		switch feature {
		case Normal:
			destroyers = append(destroyers,
				core.ResourceDestroyToZitadelDestroy(destroyCJ),
			)
			queriers = append(queriers,
				core.EnsureFuncToQueryFunc(checkDBReady),
				core.ResourceQueryToZitadelQuery(queryCJ),
			)
		case Instant:
			destroyers = append(destroyers,
				core.ResourceDestroyToZitadelDestroy(destroyJ),
			)
			queriers = append(queriers,
				core.EnsureFuncToQueryFunc(checkDBReady),
				core.ResourceQueryToZitadelQuery(queryJ),
				core.EnsureFuncToQueryFunc(getCleanupFunc(monitor, jobDef.Namespace, jobDef.Name)),
			)
		}
	}

	return func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (core.EnsureFunc, error) {
			return core.QueriersToEnsureFunc(monitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(monitor, destroyers),
		nil
}

func GetJobName(backupName string) string {
	return cronJobNamePrefix + backupName
}
