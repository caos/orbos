package api

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	orbconfig "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"strings"
)

func ParseToolset(desiredTree *tree.Tree, masterkey string) (*v1beta1.Toolset, error) {
	desiredKind := v1beta1.New(masterkey)
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	err := desiredKind.InitSecretLists(masterkey)
	return desiredKind, err
}

func SecretsFunc(orb *orbconfig.Orb) secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := ParseToolset(desiredTree, orb.Masterkey)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		return getSecretsMap(desiredKind, orb.Masterkey), nil
	}
}

func getSecretsMap(desiredKind *v1beta1.Toolset, masterkey string) map[string]*secret.Secret {
	ret := map[string]*secret.Secret{
		"argocd.sso.google.clientid":           desiredKind.Spec.Argocd.Auth.GoogleConnector.Config.ClientID,
		"argocd.sso.google.clientsecret":       desiredKind.Spec.Argocd.Auth.GoogleConnector.Config.ClientSecret,
		"argocd.sso.google.serviceaccountjson": desiredKind.Spec.Argocd.Auth.GoogleConnector.Config.ServiceAccountJSON,
		"argocd.sso.gitlab.clientid":           desiredKind.Spec.Argocd.Auth.GitlabConnector.Config.ClientID,
		"argocd.sso.gitlab.clientsecret":       desiredKind.Spec.Argocd.Auth.GitlabConnector.Config.ClientSecret,
		"argocd.sso.github.clientid":           desiredKind.Spec.Argocd.Auth.GithubConnector.Config.ClientID,
		"argocd.sso.github.clientsecret":       desiredKind.Spec.Argocd.Auth.GithubConnector.Config.ClientSecret,
		"argocd.sso.oidc.clientid":             desiredKind.Spec.Argocd.Auth.OIDC.ClientID,
		"argocd.sso.oidc.clientsecret":         desiredKind.Spec.Argocd.Auth.OIDC.ClientSecret,

		"grafana.sso.oauth.clientid":      desiredKind.Spec.Grafana.Auth.GenericOAuth.ClientID,
		"grafana.sso.oauth.clientsecret":  desiredKind.Spec.Grafana.Auth.GenericOAuth.ClientSecret,
		"grafana.sso.google.clientid":     desiredKind.Spec.Grafana.Auth.Google.ClientID,
		"grafana.sso.google.clientsecret": desiredKind.Spec.Grafana.Auth.Google.ClientSecret,
		"grafana.sso.github.clientid":     desiredKind.Spec.Grafana.Auth.Github.ClientID,
		"grafana.sso.github.clientsecret": desiredKind.Spec.Grafana.Auth.Github.ClientSecret,
		"grafana.sso.gitlab.clientid":     desiredKind.Spec.Grafana.Auth.Gitlab.ClientID,
		"grafana.sso.gitlab.clientsecret": desiredKind.Spec.Grafana.Auth.Gitlab.ClientSecret,

		"grafana.admin.username": desiredKind.Spec.Grafana.Admin.Username,
		"grafana.admin.password": desiredKind.Spec.Grafana.Admin.Password,
	}

	if desiredKind.Spec.Argocd != nil && desiredKind.Spec.Argocd.Credentials != nil {
		for _, value := range desiredKind.Spec.Argocd.Credentials {
			base := strings.Join([]string{"argocd", "credential", value.Name}, ".")

			key := strings.Join([]string{base, "username"}, ".")
			value.Username = secret.InitIfNil(value.Username, masterkey)
			ret[key] = value.Username

			key = strings.Join([]string{base, "password"}, ".")
			value.Password = secret.InitIfNil(value.Password, masterkey)
			ret[key] = value.Password

			key = strings.Join([]string{base, "certificate"}, ".")
			value.Certificate = secret.InitIfNil(value.Certificate, masterkey)
			ret[key] = value.Certificate
		}
	}

	if desiredKind.Spec.Argocd != nil && desiredKind.Spec.Argocd.Repositories != nil {
		for _, value := range desiredKind.Spec.Argocd.Repositories {
			base := strings.Join([]string{"argocd", "repository", value.Name}, ".")

			key := strings.Join([]string{base, "username"}, ".")
			value.Username = secret.InitIfNil(value.Username, masterkey)
			ret[key] = value.Username

			key = strings.Join([]string{base, "password"}, ".")
			value.Password = secret.InitIfNil(value.Password, masterkey)
			ret[key] = value.Password

			key = strings.Join([]string{base, "certificate"}, ".")
			value.Certificate = secret.InitIfNil(value.Certificate, masterkey)
			ret[key] = value.Certificate
		}
	}

	if desiredKind.Spec.Argocd != nil && desiredKind.Spec.Argocd.CustomImage != nil && desiredKind.Spec.Argocd.CustomImage.GopassStores != nil {
		for _, value := range desiredKind.Spec.Argocd.CustomImage.GopassStores {
			base := strings.Join([]string{"argocd", "gopass", value.StoreName}, ".")

			key := strings.Join([]string{base, "ssh"}, ".")
			value.SSHKey = secret.InitIfNil(value.SSHKey, masterkey)
			ret[key] = value.SSHKey

			key = strings.Join([]string{base, "gpg"}, ".")
			value.GPGKey = secret.InitIfNil(value.GPGKey, masterkey)
			ret[key] = value.GPGKey
		}
	}

	return ret
}
