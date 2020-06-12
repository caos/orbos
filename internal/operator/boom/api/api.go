package api

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"strings"
)

func ParseToolset(desiredTree *tree.Tree) (*v1beta1.Toolset, error) {
	desiredKind := &v1beta1.Toolset{}
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}

func SecretsFunc() secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
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
}

func RewriteFunc(newMasterkey string) secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := ParseToolset(desiredTree)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind
		secret.Masterkey = newMasterkey

		return getSecretsMap(desiredKind), nil
	}
}

func getSecretsMap(desiredKind *v1beta1.Toolset) map[string]*secret.Secret {
	ret := make(map[string]*secret.Secret, 0)

	if desiredKind.Spec.Grafana != nil {
		grafana := desiredKind.Spec.Grafana
		if grafana.Admin != nil {
			ret["grafana.admin.username"] = grafana.Admin.Username
			ret["grafana.admin.password"] = grafana.Admin.Password
		}

		if grafana.Auth != nil {
			if grafana.Auth.GenericOAuth != nil {
				ret["grafana.sso.oauth.clientid"] = secret.InitIfNil(grafana.Auth.GenericOAuth.ClientID)
				ret["grafana.sso.oauth.clientsecret"] = secret.InitIfNil(grafana.Auth.GenericOAuth.ClientSecret)
			}
			if grafana.Auth.Google != nil {
				ret["grafana.sso.google.clientid"] = secret.InitIfNil(grafana.Auth.Google.ClientID)
				ret["grafana.sso.google.clientsecret"] = secret.InitIfNil(grafana.Auth.Google.ClientSecret)
			}
			if grafana.Auth.Github != nil {
				ret["grafana.sso.github.clientid"] = secret.InitIfNil(grafana.Auth.Github.ClientID)
				ret["grafana.sso.github.clientsecret"] = secret.InitIfNil(grafana.Auth.Github.ClientSecret)
			}
			if grafana.Auth.Gitlab != nil {
				ret["grafana.sso.gitlab.clientid"] = secret.InitIfNil(grafana.Auth.Gitlab.ClientID)
				ret["grafana.sso.gitlab.clientsecret"] = secret.InitIfNil(grafana.Auth.Gitlab.ClientSecret)
			}
		}
	}

	if desiredKind.Spec.Argocd != nil {
		argocd := desiredKind.Spec.Argocd
		if argocd.Auth != nil {
			auth := argocd.Auth
			if auth.GoogleConnector != nil {
				ret["argocd.sso.google.clientid"] = secret.InitIfNil(argocd.Auth.GoogleConnector.Config.ClientID)
				ret["argocd.sso.google.clientsecret"] = secret.InitIfNil(argocd.Auth.GoogleConnector.Config.ClientSecret)
				ret["argocd.sso.google.serviceaccountjson"] = secret.InitIfNil(argocd.Auth.GoogleConnector.Config.ServiceAccountJSON)
			}
			if auth.GitlabConnector != nil {
				ret["argocd.sso.gitlab.clientid"] = secret.InitIfNil(argocd.Auth.GitlabConnector.Config.ClientID)
				ret["argocd.sso.gitlab.clientsecret"] = secret.InitIfNil(argocd.Auth.GitlabConnector.Config.ClientSecret)
			}

			if auth.OIDC != nil {
				ret["argocd.sso.oidc.clientid"] = secret.InitIfNil(argocd.Auth.OIDC.ClientID)
				ret["argocd.sso.oidc.clientsecret"] = secret.InitIfNil(argocd.Auth.OIDC.ClientSecret)
			}
		}
		if argocd.Credentials != nil {
			for _, value := range argocd.Credentials {
				base := strings.Join([]string{"argocd", "credential", value.Name}, ".")

				key := strings.Join([]string{base, "username"}, ".")
				value.Username = secret.InitIfNil(value.Username)
				ret[key] = value.Username

				key = strings.Join([]string{base, "password"}, ".")
				value.Password = secret.InitIfNil(value.Password)
				ret[key] = value.Password

				key = strings.Join([]string{base, "certificate"}, ".")
				value.Certificate = secret.InitIfNil(value.Certificate)
				ret[key] = value.Certificate
			}
		}
		if argocd.Repositories != nil {
			for _, value := range argocd.Repositories {
				base := strings.Join([]string{"argocd", "repository", value.Name}, ".")

				key := strings.Join([]string{base, "username"}, ".")
				value.Username = secret.InitIfNil(value.Username)
				ret[key] = value.Username

				key = strings.Join([]string{base, "password"}, ".")
				value.Password = secret.InitIfNil(value.Password)
				ret[key] = value.Password

				key = strings.Join([]string{base, "certificate"}, ".")
				value.Certificate = secret.InitIfNil(value.Certificate)
				ret[key] = value.Certificate
			}
		}

		if argocd.CustomImage != nil && argocd.CustomImage.GopassStores != nil {
			for _, value := range argocd.CustomImage.GopassStores {
				base := strings.Join([]string{"argocd", "gopass", value.StoreName}, ".")

				key := strings.Join([]string{base, "ssh"}, ".")
				value.SSHKey = secret.InitIfNil(value.SSHKey)
				ret[key] = value.SSHKey

				key = strings.Join([]string{base, "gpg"}, ".")
				value.GPGKey = secret.InitIfNil(value.GPGKey)
				ret[key] = value.GPGKey
			}
		}
	}

	return ret
}
