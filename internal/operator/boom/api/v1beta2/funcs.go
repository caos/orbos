package v1beta2

import (
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

func RewriteFunc(desiredTree *tree.Tree, newMasterkey string) (secrets map[string]*secret.Secret, err error) {
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

func getSecretsMap(desiredKind *Toolset) map[string]*secret.Secret {
	ret := make(map[string]*secret.Secret, 0)

	if desiredKind.Spec.Monitoring != nil {
		grafana := desiredKind.Spec.Monitoring
		if grafana.Admin != nil {
			ret["grafana.admin.username"] = grafana.Admin.Username
			ret["grafana.admin.password"] = grafana.Admin.Password
		}

		if grafana.Auth != nil {
			if grafana.Auth.GenericOAuth != nil {
				grafana.Auth.GenericOAuth.ClientID = secret.InitIfNil(grafana.Auth.GenericOAuth.ClientID)
				ret["grafana.sso.oauth.clientid"] = grafana.Auth.GenericOAuth.ClientID
				grafana.Auth.GenericOAuth.ClientSecret = secret.InitIfNil(grafana.Auth.GenericOAuth.ClientSecret)
				ret["grafana.sso.oauth.clientsecret"] = grafana.Auth.GenericOAuth.ClientSecret
			}
			if grafana.Auth.Google != nil {
				grafana.Auth.Google.ClientID = secret.InitIfNil(grafana.Auth.Google.ClientID)
				ret["grafana.sso.google.clientid"] = grafana.Auth.Google.ClientID
				grafana.Auth.Google.ClientSecret = secret.InitIfNil(grafana.Auth.Google.ClientSecret)
				ret["grafana.sso.google.clientsecret"] = grafana.Auth.Google.ClientSecret
			}
			if grafana.Auth.Github != nil {
				grafana.Auth.Github.ClientID = secret.InitIfNil(grafana.Auth.Github.ClientID)
				ret["grafana.sso.github.clientid"] = grafana.Auth.Github.ClientID
				grafana.Auth.Github.ClientSecret = secret.InitIfNil(grafana.Auth.Github.ClientSecret)
				ret["grafana.sso.github.clientsecret"] = grafana.Auth.Github.ClientSecret
			}
			if grafana.Auth.Gitlab != nil {
				grafana.Auth.Gitlab.ClientID = secret.InitIfNil(grafana.Auth.Gitlab.ClientID)
				ret["grafana.sso.gitlab.clientid"] = grafana.Auth.Gitlab.ClientID
				grafana.Auth.Gitlab.ClientSecret = secret.InitIfNil(grafana.Auth.Gitlab.ClientSecret)
				ret["grafana.sso.gitlab.clientsecret"] = grafana.Auth.Gitlab.ClientSecret
			}
		}
	}

	if desiredKind.Spec.Reconciling != nil {
		argocd := desiredKind.Spec.Reconciling
		if argocd.Auth != nil {
			auth := argocd.Auth
			if auth.GoogleConnector != nil {
				argocd.Auth.GoogleConnector.Config.ClientID = secret.InitIfNil(argocd.Auth.GoogleConnector.Config.ClientID)
				ret["argocd.sso.google.clientid"] = argocd.Auth.GoogleConnector.Config.ClientID
				argocd.Auth.GoogleConnector.Config.ClientSecret = secret.InitIfNil(argocd.Auth.GoogleConnector.Config.ClientSecret)
				ret["argocd.sso.google.clientsecret"] = argocd.Auth.GoogleConnector.Config.ClientSecret
				argocd.Auth.GoogleConnector.Config.ServiceAccountJSON = secret.InitIfNil(argocd.Auth.GoogleConnector.Config.ServiceAccountJSON)
				ret["argocd.sso.google.serviceaccountjson"] = argocd.Auth.GoogleConnector.Config.ServiceAccountJSON
			}
			if auth.GitlabConnector != nil {
				argocd.Auth.GitlabConnector.Config.ClientID = secret.InitIfNil(argocd.Auth.GitlabConnector.Config.ClientID)
				ret["argocd.sso.gitlab.clientid"] = argocd.Auth.GitlabConnector.Config.ClientID
				argocd.Auth.GitlabConnector.Config.ClientSecret = secret.InitIfNil(argocd.Auth.GitlabConnector.Config.ClientSecret)
				ret["argocd.sso.gitlab.clientsecret"] = argocd.Auth.GitlabConnector.Config.ClientSecret
			}
			if auth.GithubConnector != nil {
				argocd.Auth.GithubConnector.Config.ClientID = secret.InitIfNil(argocd.Auth.GithubConnector.Config.ClientID)
				ret["argocd.sso.github.clientid"] = argocd.Auth.GithubConnector.Config.ClientID
				argocd.Auth.GithubConnector.Config.ClientSecret = secret.InitIfNil(argocd.Auth.GithubConnector.Config.ClientSecret)
				ret["argocd.sso.github.clientsecret"] = argocd.Auth.GithubConnector.Config.ClientSecret
			}

			if auth.OIDC != nil {
				argocd.Auth.OIDC.ClientID = secret.InitIfNil(argocd.Auth.OIDC.ClientID)
				ret["argocd.sso.oidc.clientid"] = argocd.Auth.OIDC.ClientID
				argocd.Auth.OIDC.ClientSecret = secret.InitIfNil(argocd.Auth.OIDC.ClientSecret)
				ret["argocd.sso.oidc.clientsecret"] = argocd.Auth.OIDC.ClientSecret
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
