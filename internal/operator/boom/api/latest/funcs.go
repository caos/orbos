package latest

import (
	"strings"

	"github.com/caos/orbos/internal/operator/boom/api/latest/monitoring"
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func ParseToolset(desiredTree *tree.Tree) (*Toolset, error) {
	desiredKind := &Toolset{}
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredTree.Parsed = desiredKind
	return desiredKind, nil
}

func GetSecretsMap(desiredKind *Toolset) (
	map[string]*secret.Secret,
	map[string]*secret.Existing,
) {
	secrets := make(map[string]*secret.Secret, 0)
	existing := make(map[string]*secret.Existing, 0)

	if desiredKind.Spec.APIGateway == nil {
		desiredKind.Spec.APIGateway = &APIGateway{}
	}
	apigatewaySpec := desiredKind.Spec.APIGateway
	apigatewaySpec.InitSecrets()
	ambLicKey := "apigateway.licencekey"
	secrets[ambLicKey] = apigatewaySpec.LicenceKey
	existing[ambLicKey] = apigatewaySpec.ExistingLicenceKey

	if desiredKind.Spec.Monitoring == nil {
		desiredKind.Spec.Monitoring = &monitoring.Monitoring{}
	}
	monitoringSpec := desiredKind.Spec.Monitoring
	monitoringSpec.InitSecrets()
	monitoringAdminUser := "monitoring.admin.username"
	secrets[monitoringAdminUser] = monitoringSpec.Admin.Username
	existing[monitoringAdminUser] = monitoringSpec.Admin.ExistingUsername
	monitoringAdminPW := "monitoring.admin.password"
	secrets[monitoringAdminPW] = monitoringSpec.Admin.Password
	existing[monitoringAdminPW] = monitoringSpec.Admin.ExistingPassword
	monitoringoAuthClientIDKey := "monitoring.sso.oauth.clientid"
	secrets[monitoringoAuthClientIDKey] = monitoringSpec.Auth.GenericOAuth.ClientID
	existing[monitoringoAuthClientIDKey] = monitoringSpec.Auth.GenericOAuth.ExistingClientIDSecret
	monitoringoAuthClientIDSecKey := "monitoring.sso.oauth.clientsecret"
	secrets[monitoringoAuthClientIDSecKey] = monitoringSpec.Auth.GenericOAuth.ClientSecret
	existing[monitoringoAuthClientIDSecKey] = monitoringSpec.Auth.GenericOAuth.ExistingClientSecretSecret
	monitoringoGoogClientIDKey := "monitoring.sso.google.clientid"
	secrets[monitoringoGoogClientIDKey] = monitoringSpec.Auth.Google.ClientID
	existing[monitoringoGoogClientIDKey] = monitoringSpec.Auth.Google.ExistingClientIDSecret
	monitoringoGoogClientIDSecKey := "monitoring.sso.google.clientsecret"
	secrets[monitoringoGoogClientIDSecKey] = monitoringSpec.Auth.Google.ClientSecret
	existing[monitoringoGoogClientIDSecKey] = monitoringSpec.Auth.Google.ExistingClientSecretSecret
	monitoringoGHClientIDKey := "monitoring.sso.github.clientid"
	secrets[monitoringoGHClientIDKey] = monitoringSpec.Auth.Github.ClientID
	existing[monitoringoGHClientIDKey] = monitoringSpec.Auth.Github.ExistingClientIDSecret
	monitoringoGHClientIDSecKey := "monitoring.sso.github.clientsecret"
	secrets[monitoringoGHClientIDSecKey] = monitoringSpec.Auth.Github.ClientSecret
	existing[monitoringoGHClientIDSecKey] = monitoringSpec.Auth.Github.ExistingClientSecretSecret
	monitoringoGLClientIDKey := "monitoring.sso.gitlab.clientid"
	secrets[monitoringoGLClientIDKey] = monitoringSpec.Auth.Gitlab.ClientID
	existing[monitoringoGLClientIDKey] = monitoringSpec.Auth.Gitlab.ExistingClientIDSecret
	monitoringoGLClientIDSecKey := "monitoring.sso.gitlab.clientsecret"
	secrets[monitoringoGLClientIDSecKey] = monitoringSpec.Auth.Gitlab.ClientSecret
	existing[monitoringoGLClientIDSecKey] = monitoringSpec.Auth.Gitlab.ExistingClientSecretSecret

	if desiredKind.Spec.Reconciling == nil {
		desiredKind.Spec.Reconciling = &reconciling.Reconciling{}
	}
	reconcilingSpec := desiredKind.Spec.Reconciling
	reconcilingSpec.InitSecrets()

	reconcilingGoogClientIDKey := "reconciling.sso.google.clientid"
	secrets[reconcilingGoogClientIDKey] = reconcilingSpec.Auth.GoogleConnector.Config.ClientID
	existing[reconcilingGoogClientIDKey] = reconcilingSpec.Auth.GoogleConnector.Config.ExistingClientIDSecret
	reconcilingGoogClientIDSecKey := "reconciling.sso.google.clientsecret"
	secrets[reconcilingGoogClientIDSecKey] = reconcilingSpec.Auth.GoogleConnector.Config.ClientSecret
	existing[reconcilingGoogClientIDSecKey] = reconcilingSpec.Auth.GoogleConnector.Config.ExistingClientSecretSecret
	reconcilingGoogSAKey := "reconciling.sso.google.serviceaccountjson"
	secrets[reconcilingGoogSAKey] = reconcilingSpec.Auth.GoogleConnector.Config.ServiceAccountJSON
	existing[reconcilingGoogSAKey] = reconcilingSpec.Auth.GoogleConnector.Config.ExistingServiceAccountJSONSecret
	reconcilingGLClientIDKey := "reconciling.sso.gitlab.clientid"
	secrets[reconcilingGLClientIDKey] = reconcilingSpec.Auth.GitlabConnector.Config.ClientID
	existing[reconcilingGLClientIDKey] = reconcilingSpec.Auth.GitlabConnector.Config.ExistingClientIDSecret
	reconcilingGLClientIDSecKey := "reconciling.sso.gitlab.clientsecret"
	secrets[reconcilingGLClientIDSecKey] = reconcilingSpec.Auth.GitlabConnector.Config.ClientSecret
	existing[reconcilingGLClientIDSecKey] = reconcilingSpec.Auth.GitlabConnector.Config.ExistingClientSecretSecret
	reconcilingGHClientIDKey := "reconciling.sso.github.clientid"
	secrets[reconcilingGHClientIDKey] = reconcilingSpec.Auth.GithubConnector.Config.ClientID
	existing[reconcilingGHClientIDKey] = reconcilingSpec.Auth.GithubConnector.Config.ExistingClientIDSecret
	reconcilingGHClientIDSecKey := "reconciling.sso.github.clientsecret"
	secrets[reconcilingGHClientIDSecKey] = reconcilingSpec.Auth.GithubConnector.Config.ClientSecret
	existing[reconcilingGHClientIDSecKey] = reconcilingSpec.Auth.GithubConnector.Config.ExistingClientSecretSecret
	reconcilingOIDCClientIDKey := "reconciling.sso.oidc.clientid"
	secrets[reconcilingOIDCClientIDKey] = reconcilingSpec.Auth.OIDC.ClientID
	existing[reconcilingOIDCClientIDKey] = reconcilingSpec.Auth.OIDC.ExistingClientIDSecret
	reconcilingOIDCClientIDSecKey := "reconciling.sso.oidc.clientsecret"
	secrets[reconcilingOIDCClientIDSecKey] = reconcilingSpec.Auth.OIDC.ClientSecret
	existing[reconcilingOIDCClientIDSecKey] = reconcilingSpec.Auth.OIDC.ExistingClientSecretSecret

	if reconcilingSpec.Credentials != nil {
		for _, value := range reconcilingSpec.Credentials {
			base := strings.Join([]string{"reconciling", "credential", value.Name}, ".")

			key := strings.Join([]string{base, "username"}, ".")
			if value.Username == nil {
				value.Username = &secret.Secret{}
			}
			secrets[key] = value.Username

			key = strings.Join([]string{base, "password"}, ".")
			if value.Password == nil {
				value.Password = &secret.Secret{}
			}
			secrets[key] = value.Password

			key = strings.Join([]string{base, "certificate"}, ".")
			if value.Certificate == nil {
				value.Certificate = &secret.Secret{}
			}
			secrets[key] = value.Certificate
		}
	}
	if reconcilingSpec.Repositories != nil {
		for _, value := range reconcilingSpec.Repositories {
			base := strings.Join([]string{"reconciling", "repository", value.Name}, ".")

			key := strings.Join([]string{base, "username"}, ".")
			if value.Username == nil {
				value.Username = &secret.Secret{}
			}
			secrets[key] = value.Username

			key = strings.Join([]string{base, "password"}, ".")
			if value.Password == nil {
				value.Password = &secret.Secret{}
			}
			secrets[key] = value.Password

			key = strings.Join([]string{base, "certificate"}, ".")
			if value.Certificate == nil {
				value.Certificate = &secret.Secret{}
			}
			secrets[key] = value.Certificate
		}
	}

	if reconcilingSpec.CustomImage != nil && reconcilingSpec.CustomImage.GopassStores != nil {
		for _, value := range reconcilingSpec.CustomImage.GopassStores {
			base := strings.Join([]string{"reconciling", "gopass", value.StoreName}, ".")

			key := strings.Join([]string{base, "ssh"}, ".")
			if value.SSHKey == nil {
				value.SSHKey = &secret.Secret{}
			}
			secrets[key] = value.SSHKey

			key = strings.Join([]string{base, "gpg"}, ".")
			if value.GPGKey == nil {
				value.GPGKey = &secret.Secret{}
			}
			secrets[key] = value.GPGKey
		}
	}

	return secrets, existing
}
