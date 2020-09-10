package v1beta2

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
	"strings"
)

func ParseToolset(desiredTree *tree.Tree) (*Toolset, error) {
	desiredKind := &Toolset{}
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}

func SecretsFunc(desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
	defer func() {
		err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
	}()

	desiredKind, err := ParseToolset(desiredTree)
	if err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredTree.Parsed = desiredKind

	return getSecretsMap(desiredKind), nil
}

func getSecretsMap(desiredKind *Toolset) map[string]*secret.Secret {
	ret := make(map[string]*secret.Secret, 0)

	if desiredKind.Spec.Monitoring == nil {
		desiredKind.Spec.Monitoring = &monitoring.Monitoring{}
	}
	grafanaSpec := desiredKind.Spec.Monitoring
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

	if desiredKind.Spec.Reconciling == nil {
		desiredKind.Spec.Reconciling = &reconciling.Reconciling{}
	}
	argocdSpec := desiredKind.Spec.Reconciling
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
