package admin

import (
	"strings"

	"github.com/caos/orbos/pkg/helper"

	"github.com/caos/orbos/internal/operator/boom/api/latest/monitoring/admin"
	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana/helm"
	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana/info"
	"github.com/caos/orbos/internal/operator/boom/application/resources"
	"github.com/caos/orbos/internal/operator/boom/labels"
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

	if adminSpec == nil {
		return secrets
	}

	if !helper.IsExistentClientSecret(adminSpec.ExistingSecret) {
		if adminSpec.Username.Value == "" && adminSpec.Password.Value == "" {
			return secrets
		}

		data := make(map[string]string, 0)
		if adminSpec.Username.Value != "" {
			key := getUserKey()
			data[key] = adminSpec.Username.Value
		}
		if adminSpec.Password.Value != "" {
			key := getPasswordKey()
			data[key] = adminSpec.Password.Value

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
	if helper.IsExistentClientSecret(adminSpec.ExistingSecret) {
		return &helm.Admin{
			ExistingSecret: adminSpec.ExistingSecret.Name,
			UserKey:        adminSpec.ExistingSecret.IDKey,
			PasswordKey:    adminSpec.ExistingSecret.SecretKey,
		}
	}
	if len(GetSecrets(adminSpec)) == 0 {
		return nil
	}

	return &helm.Admin{
		ExistingSecret: getSecretName(),
		UserKey:        getUserKey(),
		PasswordKey:    getPasswordKey(),
	}

}
