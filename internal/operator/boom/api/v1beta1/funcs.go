package v1beta1

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
	"strings"
)

func ParseToolset(desiredTree *tree.Tree) (*Toolset, map[string]*secret.Secret, error) {
	desiredKind := &Toolset{}
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, GetSecretsMap(desiredKind), nil
}

func GetSecretsMap(desiredKind *Toolset) map[string]*secret.Secret {
	ret := make(map[string]*secret.Secret, 0)

	if desiredKind.Spec.Grafana == nil {
		desiredKind.Spec.Grafana = &grafana.Grafana{}
	}
	grafanaSpec := desiredKind.Spec.Grafana
	grafanaSpec.InitSecrets()

	ret["grafana.admin.username"] = grafanaSpec.Admin.Username
	ret["grafana.admin.password"] = grafanaSpec.Admin.Password
	ret["grafana.sso.oauth.clientid"] = grafanaSpec.Auth.GenericOAuth.ClientID
	ret["grafana.sso.oauth.clientsecret"] = grafanaSpec.Auth.GenericOAuth.ClientSecret
	ret["grafana.sso.google.clientid"] = grafanaSpec.Auth.Google.ClientID
	ret["grafana.sso.google.clientsecret"] = grafanaSpec.Auth.Google.ClientSecret
	ret["grafana.sso.github.clientid"] = grafanaSpec.Auth.Github.ClientID
	ret["grafana.sso.github.clientsecret"] = grafanaSpec.Auth.Github.ClientSecret
	ret["grafana.sso.gitlab.clientid"] = grafanaSpec.Auth.Gitlab.ClientID
	ret["grafana.sso.gitlab.clientsecret"] = grafanaSpec.Auth.Gitlab.ClientSecret

	if desiredKind.Spec.Argocd == nil {
		desiredKind.Spec.Argocd = &argocd.Argocd{}
	}
	argocdSpec := desiredKind.Spec.Argocd
	argocdSpec.InitSecrets()

	ret["argocd.sso.google.clientid"] = argocdSpec.Auth.GoogleConnector.Config.ClientID
	ret["argocd.sso.google.clientsecret"] = argocdSpec.Auth.GoogleConnector.Config.ClientSecret
	ret["argocd.sso.google.serviceaccountjson"] = argocdSpec.Auth.GoogleConnector.Config.ServiceAccountJSON
	ret["argocd.sso.gitlab.clientid"] = argocdSpec.Auth.GitlabConnector.Config.ClientID
	ret["argocd.sso.gitlab.clientsecret"] = argocdSpec.Auth.GitlabConnector.Config.ClientSecret
	ret["argocd.sso.github.clientid"] = argocdSpec.Auth.GithubConnector.Config.ClientID
	ret["argocd.sso.github.clientsecret"] = argocdSpec.Auth.GithubConnector.Config.ClientSecret
	ret["argocd.sso.oidc.clientid"] = argocdSpec.Auth.OIDC.ClientID
	ret["argocd.sso.oidc.clientsecret"] = argocdSpec.Auth.OIDC.ClientSecret

	if argocdSpec.Credentials != nil {
		for _, value := range argocdSpec.Credentials {
			base := strings.Join([]string{"argocd", "credential", value.Name}, ".")

			key := strings.Join([]string{base, "username"}, ".")
			if value.Username == nil {
				value.Username = &secret.Secret{}
			}
			ret[key] = value.Username

			key = strings.Join([]string{base, "password"}, ".")
			if value.Password == nil {
				value.Password = &secret.Secret{}
			}
			ret[key] = value.Password

			key = strings.Join([]string{base, "certificate"}, ".")
			if value.Certificate == nil {
				value.Certificate = &secret.Secret{}
			}
			ret[key] = value.Certificate
		}
	}
	if argocdSpec.Repositories != nil {
		for _, value := range argocdSpec.Repositories {
			base := strings.Join([]string{"argocd", "repository", value.Name}, ".")

			key := strings.Join([]string{base, "username"}, ".")
			if value.Username == nil {
				value.Username = &secret.Secret{}
			}
			ret[key] = value.Username

			key = strings.Join([]string{base, "password"}, ".")
			if value.Password == nil {
				value.Password = &secret.Secret{}
			}
			ret[key] = value.Password

			key = strings.Join([]string{base, "certificate"}, ".")
			if value.Certificate == nil {
				value.Certificate = &secret.Secret{}
			}
			ret[key] = value.Certificate
		}
	}

	if argocdSpec.CustomImage != nil && argocdSpec.CustomImage.GopassStores != nil {
		for _, value := range argocdSpec.CustomImage.GopassStores {
			base := strings.Join([]string{"argocd", "gopass", value.StoreName}, ".")

			key := strings.Join([]string{base, "ssh"}, ".")
			if value.SSHKey == nil {
				value.SSHKey = &secret.Secret{}
			}
			ret[key] = value.SSHKey

			key = strings.Join([]string{base, "gpg"}, ".")
			if value.GPGKey == nil {
				value.GPGKey = &secret.Secret{}
			}
			ret[key] = value.GPGKey
		}
	}

	return ret
}
