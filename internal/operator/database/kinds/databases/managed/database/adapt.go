package database

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	deployName string,
	containerName string,
	certsDir string,
	userName string,
) (
	core.QueryFunc,
	core.DestroyFunc,
	error,
) {
	cmdSql := fmt.Sprintf("cockroach sql --certs-dir=%s", certsDir)

	createSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s ", userName)

	deleteSql := fmt.Sprintf("DROP DATABASE IF EXISTS %s", userName)

	ensureDatabase := func(k8sClient kubernetes.ClientInt) error {
		return k8sClient.ExecInPodOfDeployment(namespace, deployName, containerName, fmt.Sprintf("%s -e '%s;'", cmdSql, createSql))
	}

	destroyDatabase := func(k8sClient kubernetes.ClientInt) error {
		return k8sClient.ExecInPodOfDeployment(namespace, deployName, containerName, fmt.Sprintf("%s -e '%s;'", cmdSql, deleteSql))
	}

	queriers := []core.QueryFunc{
		core.EnsureFuncToQueryFunc(ensureDatabase),
	}

	destroyers := []core.DestroyFunc{
		destroyDatabase,
	}

	return func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (core.EnsureFunc, error) {
			return core.QueriersToEnsureFunc(monitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(monitor, destroyers),
		nil
}
