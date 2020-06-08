package helper

import (
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/utils/clientgo"
)

func IsExistentSecret(secret *secret.Secret, existent *secret.Existing) bool {
	if (secret == nil || secret.Value == "") && existent != nil && (existent.Name != "" && existent.Key != "") {
		return true
	}
	return false
}

func IsExistentClientSecret(existent *secret.ExistingIDSecret) bool {
	if existent != nil && (existent.Name != "" && existent.IDKey != "" && existent.SecretKey != "") {
		return true
	}
	return false
}

func IsCrdSecret(secret *secret.Secret, existent *secret.Existing) bool {
	if (secret != nil && secret.Value != "") && (existent == nil || (existent.Name == "" || existent.Key == "")) {
		return true
	}
	return false
}

func GetSecretValue(secret *secret.Secret, existing *secret.Existing) (string, error) {
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
