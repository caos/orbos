package v1beta2

import (
	orbconfig "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
	"strings"
)

func ParseToolset(desiredTree *tree.Tree, masterkey string) (*Toolset, error) {
	desiredKind := New(masterkey)
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	err := desiredKind.InitSecretLists(masterkey)
	return desiredKind, err
}

func SecretsFunc(orb *orbconfig.Orb, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
	defer func() {
		err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
	}()

	desiredKind, err := ParseToolset(desiredTree, orb.Masterkey)
	if err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredTree.Parsed = desiredKind

	return GetSecretsMap(desiredKind, orb.Masterkey), nil
}

func RewriteFunc(orb *orbconfig.Orb, desiredTree *tree.Tree, masterkeyNew string) (secrets map[string]*secret.Secret, err error) {
	defer func() {
		err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
	}()

	desiredKind, err := ParseToolset(desiredTree, orb.Masterkey)
	if err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredKind = ReplaceMasterkey(desiredKind, masterkeyNew)
	desiredTree.Parsed = desiredKind

	return GetSecretsMap(desiredKind, orb.Masterkey), nil
}

func GetSecretsMap(desiredKind *Toolset, masterkey string) map[string]*secret.Secret {
	ret := map[string]*secret.Secret{
		"reconciling.sso.google.clientid":           desiredKind.Spec.Reconciling.Auth.GoogleConnector.Config.ClientID,
		"reconciling.sso.google.clientsecret":       desiredKind.Spec.Reconciling.Auth.GoogleConnector.Config.ClientSecret,
		"reconciling.sso.google.serviceaccountjson": desiredKind.Spec.Reconciling.Auth.GoogleConnector.Config.ServiceAccountJSON,
		"reconciling.sso.gitlab.clientid":           desiredKind.Spec.Reconciling.Auth.GitlabConnector.Config.ClientID,
		"reconciling.sso.gitlab.clientsecret":       desiredKind.Spec.Reconciling.Auth.GitlabConnector.Config.ClientSecret,
		"reconciling.sso.github.clientid":           desiredKind.Spec.Reconciling.Auth.GithubConnector.Config.ClientID,
		"reconciling.sso.github.clientsecret":       desiredKind.Spec.Reconciling.Auth.GithubConnector.Config.ClientSecret,
		"reconciling.sso.oidc.clientid":             desiredKind.Spec.Reconciling.Auth.OIDC.ClientID,
		"reconciling.sso.oidc.clientsecret":         desiredKind.Spec.Reconciling.Auth.OIDC.ClientSecret,

		"monitoring.sso.oauth.clientid":      desiredKind.Spec.Monitoring.Auth.GenericOAuth.ClientID,
		"monitoring.sso.oauth.clientsecret":  desiredKind.Spec.Monitoring.Auth.GenericOAuth.ClientSecret,
		"monitoring.sso.google.clientid":     desiredKind.Spec.Monitoring.Auth.Google.ClientID,
		"monitoring.sso.google.clientsecret": desiredKind.Spec.Monitoring.Auth.Google.ClientSecret,
		"monitoring.sso.github.clientid":     desiredKind.Spec.Monitoring.Auth.Github.ClientID,
		"monitoring.sso.github.clientsecret": desiredKind.Spec.Monitoring.Auth.Github.ClientSecret,
		"monitoring.sso.gitlab.clientid":     desiredKind.Spec.Monitoring.Auth.Gitlab.ClientID,
		"monitoring.sso.gitlab.clientsecret": desiredKind.Spec.Monitoring.Auth.Gitlab.ClientSecret,

		"grafana.admin.username": desiredKind.Spec.Monitoring.Admin.Username,
		"grafana.admin.password": desiredKind.Spec.Monitoring.Admin.Password,
	}

	if desiredKind.Spec.Reconciling != nil && desiredKind.Spec.Reconciling.Credentials != nil {
		for _, value := range desiredKind.Spec.Reconciling.Credentials {
			base := strings.Join([]string{"reconciling", "credential", value.Name}, ".")

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

	if desiredKind.Spec.Reconciling != nil && desiredKind.Spec.Reconciling.Repositories != nil {
		for _, value := range desiredKind.Spec.Reconciling.Repositories {
			base := strings.Join([]string{"reconciling", "repository", value.Name}, ".")

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

	if desiredKind.Spec.Reconciling != nil && desiredKind.Spec.Reconciling.CustomImage != nil && desiredKind.Spec.Reconciling.CustomImage.GopassStores != nil {
		for _, value := range desiredKind.Spec.Reconciling.CustomImage.GopassStores {
			base := strings.Join([]string{"reconciling", "gopass", value.StoreName}, ".")

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
