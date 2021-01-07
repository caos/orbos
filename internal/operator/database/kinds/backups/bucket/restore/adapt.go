package restore

import (
	"time"

	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/job"
	"github.com/caos/orbos/pkg/labels"
	corev1 "k8s.io/api/core/v1"
)

const (
	Instant                          = "restore"
	defaultMode                      = int32(256)
	certPath                         = "/cockroach/cockroach-certs"
	secretPath                       = "/secrets/sa.json"
	jobPrefix                        = "backup-"
	jobSuffix                        = "-restore"
	image                            = "cockroachdb/cockroach:v20.2.3"
	internalSecretName               = "client-certs"
	rootSecretName                   = "cockroachdb.client.root"
	timeout            time.Duration = 60
	saJsonBase64Env                  = "SAJSON"
)

func AdaptFunc(
	monitor mntr.Monitor,
	backupName string,
	namespace string,
	componentLabels *labels.Component,
	bucketName string,
	timestamp string,
	nodeselector map[string]string,
	tolerations []corev1.Toleration,
	checkDBReady core.EnsureFunc,
	secretName string,
	secretKey string,
	dbURL string,
	dbPort int32,
) (
	queryFunc core.QueryFunc,
	destroyFunc core.DestroyFunc,
	err error,
) {

	jobName := jobPrefix + backupName + jobSuffix
	command := getCommand(
		timestamp,
		bucketName,
		backupName,
		certPath,
		secretPath,
		dbURL,
		dbPort,
	)

	jobdef := getJob(
		namespace,
		labels.MustForName(componentLabels, GetJobName(backupName)),
		nodeselector,
		tolerations,
		secretName,
		secretKey,
		command)

	destroyJ, err := job.AdaptFuncToDestroy(jobName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyers := []core.DestroyFunc{
		core.ResourceDestroyToZitadelDestroy(destroyJ),
	}

	queryJ, err := job.AdaptFuncToEnsure(jobdef)
	if err != nil {
		return nil, nil, err
	}

	queriers := []core.QueryFunc{
		core.EnsureFuncToQueryFunc(checkDBReady),
		core.ResourceQueryToZitadelQuery(queryJ),
		//core.EnsureFuncToQueryFunc(getCleanupFunc(monitor, jobdef.Namespace, jobdef.Name)),
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
