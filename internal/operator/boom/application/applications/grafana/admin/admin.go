package admin

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/admin"
	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana/helm"
	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana/info"
	"github.com/caos/orbos/internal/operator/boom/application/resources"
	"github.com/caos/orbos/internal/operator/boom/labels"
	helper2 "github.com/caos/orbos/internal/utils/helper"
	"strings"
)

func getSecretName() string {
	return strings.Join([]string{"grafana", "admin"}, "-")
}

func getUserKey() string {
	return "username"
}

func getPasswordKey() string {
	return "password"
}

func GetSecrets(adminSpec *admin.Admin) []interface{} {
	namespace := "caos-system"

	secrets := make([]interface{}, 0)

	if !helper2.IsExistentClientSecret(adminSpec.ExistingSecret) {
		data := map[string]string{
			getUserKey():     adminSpec.Username.Value,
			getPasswordKey(): adminSpec.Password.Value,
		}

		conf := &resources.SecretConfig{
			Name:      getSecretName(),
			Namespace: namespace,
			Labels:    labels.GetAllApplicationLabels(info.GetName()),
			Data:      data,
		}
		secretRes := resources.NewSecret(conf)
		secrets = append(secrets, secretRes)
	}
	return secrets
}

func GetConfig(adminSpec *admin.Admin) *helm.Admin {
	if helper2.IsExistentClientSecret(adminSpec.ExistingSecret) {

		return &helm.Admin{
			ExistingSecret: adminSpec.ExistingSecret.Name,
			UserKey:        adminSpec.ExistingSecret.IDKey,
			PasswordKey:    adminSpec.ExistingSecret.SecretKey,
		}
	}

	return &helm.Admin{
		ExistingSecret: getSecretName(),
		UserKey:        getUserKey(),
		PasswordKey:    getPasswordKey(),
	}

}
