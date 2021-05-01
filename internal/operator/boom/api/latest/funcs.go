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
	ambassadorSpec := desiredKind.Spec.APIGateway
	ambassadorSpec.InitSecrets()
	ambLicKey := "ambassador.licencekey"
	secrets[ambLicKey] = ambassadorSpec.LicenceKey
	existing[ambLicKey] = ambassadorSpec.ExistingLicenceKey

	if desiredKind.Spec.Monitoring == nil {
		desiredKind.Spec.Monitoring = &monitoring.Monitoring{}
	}
	grafanaSpec := desiredKind.Spec.Monitoring
	grafanaSpec.InitSecrets()
	grafAdminUser := "grafana.admin.username"
	secrets[grafAdminUser] = grafanaSpec.Admin.Username
	existing[grafAdminUser] = grafanaSpec.Admin.ExistingUsername
	grafAdminPW := "grafana.admin.password"
	secrets[grafAdminPW] = grafanaSpec.Admin.Password
	existing[grafAdminPW] = grafanaSpec.Admin.ExistingPassword
	grafoAuthClientIDKey := "grafana.sso.oauth.clientid"
	secrets[grafoAuthClientIDKey] = grafanaSpec.Auth.GenericOAuth.ClientID
	existing[grafoAuthClientIDKey] = grafanaSpec.Auth.GenericOAuth.ExistingClientIDSecret
	grafoAuthClientIDSecKey := "grafana.sso.oauth.clientsecret"
	secrets[grafoAuthClientIDSecKey] = grafanaSpec.Auth.GenericOAuth.ClientSecret
	existing[grafoAuthClientIDSecKey] = grafanaSpec.Auth.GenericOAuth.ExistingClientSecretSecret
	grafoGoogClientIDKey := "grafana.sso.google.clientid"
	secrets[grafoGoogClientIDKey] = grafanaSpec.Auth.Google.ClientID
	existing[grafoGoogClientIDKey] = grafanaSpec.Auth.Google.ExistingClientIDSecret
	grafoGoogClientIDSecKey := "grafana.sso.google.clientsecret"
	secrets[grafoGoogClientIDSecKey] = grafanaSpec.Auth.Google.ClientSecret
	existing[grafoGoogClientIDSecKey] = grafanaSpec.Auth.Google.ExistingClientSecretSecret
	grafoGHClientIDKey := "grafana.sso.github.clientid"
	secrets[grafoGHClientIDKey] = grafanaSpec.Auth.Github.ClientID
	existing[grafoGHClientIDKey] = grafanaSpec.Auth.Github.ExistingClientIDSecret
	grafoGHClientIDSecKey := "grafana.sso.github.clientsecret"
	secrets[grafoGHClientIDSecKey] = grafanaSpec.Auth.Github.ClientSecret
	existing[grafoGHClientIDSecKey] = grafanaSpec.Auth.Github.ExistingClientSecretSecret
	grafoGLClientIDKey := "grafana.sso.gitlab.clientid"
	secrets[grafoGLClientIDKey] = grafanaSpec.Auth.Gitlab.ClientID
	existing[grafoGLClientIDKey] = grafanaSpec.Auth.Gitlab.ExistingClientIDSecret
	grafoGLClientIDSecKey := "grafana.sso.gitlab.clientsecret"
	secrets[grafoGLClientIDSecKey] = grafanaSpec.Auth.Gitlab.ClientSecret
	existing[grafoGLClientIDSecKey] = grafanaSpec.Auth.Gitlab.ExistingClientSecretSecret

	if desiredKind.Spec.Reconciling == nil {
		desiredKind.Spec.Reconciling = &reconciling.Reconciling{}
	}
	argocdSpec := desiredKind.Spec.Reconciling
	argocdSpec.InitSecrets()

	argoGoogClientIDKey := "argocd.sso.google.clientid"
	secrets[argoGoogClientIDKey] = argocdSpec.Auth.GoogleConnector.Config.ClientID
	existing[argoGoogClientIDKey] = argocdSpec.Auth.GoogleConnector.Config.ExistingClientIDSecret
	argoGoogClientIDSecKey := "argocd.sso.google.clientsecret"
	secrets[argoGoogClientIDSecKey] = argocdSpec.Auth.GoogleConnector.Config.ClientSecret
	existing[argoGoogClientIDSecKey] = argocdSpec.Auth.GoogleConnector.Config.ExistingClientSecretSecret
	argoGoogSAKey := "argocd.sso.google.serviceaccountjson"
	secrets[argoGoogSAKey] = argocdSpec.Auth.GoogleConnector.Config.ServiceAccountJSON
	existing[argoGoogSAKey] = argocdSpec.Auth.GoogleConnector.Config.ExistingServiceAccountJSONSecret
	argoGLClientIDKey := "argocd.sso.gitlab.clientid"
	secrets[argoGLClientIDKey] = argocdSpec.Auth.GitlabConnector.Config.ClientID
	existing[argoGLClientIDKey] = argocdSpec.Auth.GitlabConnector.Config.ExistingClientIDSecret
	argoGLClientIDSecKey := "argocd.sso.gitlab.clientsecret"
	secrets[argoGLClientIDSecKey] = argocdSpec.Auth.GitlabConnector.Config.ClientSecret
	existing[argoGLClientIDSecKey] = argocdSpec.Auth.GitlabConnector.Config.ExistingClientSecretSecret
	argoGHClientIDKey := "argocd.sso.github.clientid"
	secrets[argoGHClientIDKey] = argocdSpec.Auth.GithubConnector.Config.ClientID
	existing[argoGHClientIDKey] = argocdSpec.Auth.GithubConnector.Config.ExistingClientIDSecret
	argoGHClientIDSecKey := "argocd.sso.github.clientsecret"
	secrets[argoGHClientIDSecKey] = argocdSpec.Auth.GithubConnector.Config.ClientSecret
	existing[argoGHClientIDSecKey] = argocdSpec.Auth.GithubConnector.Config.ExistingClientSecretSecret
	argoOIDCClientIDKey := "argocd.sso.oidc.clientid"
	secrets[argoOIDCClientIDKey] = argocdSpec.Auth.OIDC.ClientID
	existing[argoOIDCClientIDKey] = argocdSpec.Auth.OIDC.ExistingClientIDSecret
	argoOIDCClientIDSecKey := "argocd.sso.oidc.clientsecret"
	secrets[argoOIDCClientIDSecKey] = argocdSpec.Auth.OIDC.ClientSecret
	existing[argoOIDCClientIDSecKey] = argocdSpec.Auth.OIDC.ExistingClientSecretSecret

	if argocdSpec.Credentials != nil {
		for _, value := range argocdSpec.Credentials {
			base := strings.Join([]string{"argocd", "credential", value.Name}, ".")

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
	if argocdSpec.Repositories != nil {
		for _, value := range argocdSpec.Repositories {
			base := strings.Join([]string{"argocd", "repository", value.Name}, ".")

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

	if argocdSpec.CustomImage != nil && argocdSpec.CustomImage.GopassStores != nil {
		for _, value := range argocdSpec.CustomImage.GopassStores {
			base := strings.Join([]string{"argocd", "gopass", value.StoreName}, ".")

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
