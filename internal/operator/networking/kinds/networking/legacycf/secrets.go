package legacycf

import (
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/config"
	"github.com/caos/orbos/pkg/secret"
)

func getSecretsMap(desiredKind *Desired) (map[string]*secret.Secret, map[string]*secret.Existing) {
	secrets := map[string]*secret.Secret{}
	existing := map[string]*secret.Existing{}
	if desiredKind.Spec == nil {
		desiredKind.Spec = &config.ExternalConfig{}
	}

	if desiredKind.Spec.Credentials == nil {
		desiredKind.Spec.Credentials = &config.Credentials{}
	}

	if desiredKind.Spec.Credentials.User == nil {
		desiredKind.Spec.Credentials.User = &secret.Secret{}
	}
	if desiredKind.Spec.Credentials.ExistingUser == nil {
		desiredKind.Spec.Credentials.ExistingUser = &secret.Existing{}
	}

	if desiredKind.Spec.Credentials.APIKey == nil {
		desiredKind.Spec.Credentials.APIKey = &secret.Secret{}
	}
	if desiredKind.Spec.Credentials.ExistingAPIKey == nil {
		desiredKind.Spec.Credentials.ExistingAPIKey = &secret.Existing{}
	}

	if desiredKind.Spec.Credentials.UserServiceKey == nil {
		desiredKind.Spec.Credentials.UserServiceKey = &secret.Secret{}
	}
	if desiredKind.Spec.Credentials.ExistingUserServiceKey == nil {
		desiredKind.Spec.Credentials.ExistingUserServiceKey = &secret.Existing{}
	}

	userKey := "credentials.user"
	secrets[userKey] = desiredKind.Spec.Credentials.User
	existing[userKey] = desiredKind.Spec.Credentials.ExistingUser
	apiKey := "credentials.apikey"
	secrets[apiKey] = desiredKind.Spec.Credentials.APIKey
	existing[apiKey] = desiredKind.Spec.Credentials.ExistingAPIKey
	svcKey := "credentials.userservicekey"
	secrets[svcKey] = desiredKind.Spec.Credentials.UserServiceKey
	existing[svcKey] = desiredKind.Spec.Credentials.ExistingUserServiceKey

	return secrets, existing
}
