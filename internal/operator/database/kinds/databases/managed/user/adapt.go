package user

import (
	"fmt"
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/certificate"
	"github.com/caos/orbos/mntr"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	deployName string,
	containerName string,
	certsDir string,
	userName string,
	password string,
	labels map[string]string,
) (
	core.QueryFunc,
	core.DestroyFunc,
	error,
) {
	cmdSql := fmt.Sprintf("cockroach sql --certs-dir=%s", certsDir)

	createSql := fmt.Sprintf("CREATE USER IF NOT EXISTS %s ", userName)
	if password != "" {
		createSql = fmt.Sprintf("%s WITH PASSWORD %s", createSql, password)
	}

	deleteSql := fmt.Sprintf("DROP USER IF EXISTS %s", userName)

	_, _, addUserFunc, deleteUserFunc, _, err := certificate.AdaptFunc(monitor, namespace, labels, "")
	if err != nil {
		return nil, nil, err
	}

	addUser, err := addUserFunc(userName)
	if err != nil {
		return nil, nil, err
	}
	ensureUser := func(k8sClient *kubernetes2.Client) error {
		return k8sClient.ExecInPodOfDeployment(namespace, deployName, containerName, fmt.Sprintf("%s -e '%s;'", cmdSql, createSql))
	}

	deleteUser, err := deleteUserFunc(userName)
	if err != nil {
		return nil, nil, err
	}
	destoryUser := func(k8sClient *kubernetes2.Client) error {
		return k8sClient.ExecInPodOfDeployment(namespace, deployName, containerName, fmt.Sprintf("%s -e '%s;'", cmdSql, deleteSql))
	}

	queriers := []core.QueryFunc{
		addUser,
		core.EnsureFuncToQueryFunc(ensureUser),
	}

	destroyers := []core.DestroyFunc{
		destoryUser,
		deleteUser,
	}

	return func(k8sClient *kubernetes2.Client, queried map[string]interface{}) (core.EnsureFunc, error) {
			return core.QueriersToEnsureFunc(monitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(monitor, destroyers),
		nil
}
