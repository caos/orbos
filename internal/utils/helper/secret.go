package helper

import (
	"github.com/caos/orbos/internal/utils/clientgo"
	secret2 "github.com/caos/orbos/pkg/secret"
)

func IsExistentSecret(secret *secret2.Secret, existent *secret2.Existing) bool {
	if (secret == nil || secret.Value == "") && existent != nil && (existent.Name != "" && existent.Key != "") {
		return true
	}
	return false
}

func IsExistentClientSecret(existent *secret2.ExistingIDSecret) bool {
	if existent != nil && (existent.Name != "" && existent.IDKey != "" && existent.SecretKey != "") {
		return true
	}
	return false
}

func IsCrdSecret(secret *secret2.Secret, existent *secret2.Existing) bool {
	if (secret != nil && secret.Value != "") && (existent == nil || (existent.Name == "" || existent.Key == "")) {
		return true
	}
	return false
}

func GetSecretValue(secret *secret2.Secret, existing *secret2.Existing) (string, error) {
	if IsExistentSecret(secret, existing) {
		secret, err := clientgo.GetSecret(existing.Name, "caos-system")
		if err != nil {
			return "", err
		}

		return string(secret.Data[existing.Key]), nil
	} else if IsCrdSecret(secret, existing) {
		return secret.Value, nil
	}

	return "", nil
}
