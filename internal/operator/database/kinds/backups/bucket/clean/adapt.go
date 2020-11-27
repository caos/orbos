package clean

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/job"
	corev1 "k8s.io/api/core/v1"
	"time"
)

const (
	defaultMode                      = int32(256)
	certPath                         = "/cockroach/cockroach-certs"
	secretPath                       = "/secrets/sa.json"
	internalSecretName               = "client-certs"
	image                            = "ghcr.io/caos/crbackup"
	rootSecretName                   = "cockroachdb.client.root"
	jobPrefix                        = "backup-"
	jobSuffix                        = "-clean"
	timeout            time.Duration = 60
)

func AdaptFunc(
	monitor mntr.Monitor,
	backupName string,
	namespace string,
	labels map[string]string,
	databases []string,
	nodeselector map[string]string,
	tolerations []corev1.Toleration,
	checkDBReady core.EnsureFunc,
	secretName string,
	secretKey string,
	version string,
) (
	queryFunc core.QueryFunc,
	destroyFunc core.DestroyFunc,
	err error,
) {

	command := getCommand(databases)

	jobDef := getJob(
		namespace,
		labels,
		GetJobName(backupName),
		nodeselector,
		tolerations,
		secretName,
		secretKey,
		version,
		command)

	destroyJ, err := job.AdaptFuncToDestroy(jobDef.Namespace, jobDef.Name)
	if err != nil {
		return nil, nil, err
	}

	destroyers := []core.DestroyFunc{
		core.ResourceDestroyToZitadelDestroy(destroyJ),
	}

	queryJ, err := job.AdaptFuncToEnsure(jobDef)
	if err != nil {
		return nil, nil, err
	}

	queriers := []core.QueryFunc{
		core.EnsureFuncToQueryFunc(checkDBReady),
		core.ResourceQueryToZitadelQuery(queryJ),
		core.EnsureFuncToQueryFunc(getCleanupFunc(monitor, jobDef.Namespace, jobDef.Name)),
	}

	return func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (core.EnsureFunc, error) {
			return core.QueriersToEnsureFunc(monitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(monitor, destroyers),
		nil
}

func GetJobName(backupName string) string {
	return jobPrefix + backupName + jobSuffix
}
