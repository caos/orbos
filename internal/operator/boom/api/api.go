package api

import (
	"github.com/caos/orbiter/internal/operator/boom/api/v1beta1"
	orbconfig "github.com/caos/orbiter/internal/orb"
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/internal/tree"
	"github.com/caos/orbiter/mntr"
	"github.com/pkg/errors"
)

func parseToolset(desiredTree *tree.Tree, masterkey string) (*v1beta1.Toolset, error) {
	desiredKind := &v1beta1.Toolset{
		Spec: &v1beta1.ToolsetSpec{
			Grafana: &v1beta1.Grafana{
				Admin: &v1beta1.Admin{
					Username: &secret.Secret{Masterkey: masterkey},
					Password: &secret.Secret{Masterkey: masterkey},
				},
				Auth: &v1beta1.GrafanaAuth{
					Google: &v1beta1.GrafanaGoogleAuth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					Github: &v1beta1.GrafanaGithubAuth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					Gitlab: &v1beta1.GrafanaGitlabAuth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					GenericOAuth: &v1beta1.GrafanaGenericOAuth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
				},
			},
			Argocd: &v1beta1.Argocd{
				Auth: &v1beta1.ArgocdAuth{
					OIDC: &v1beta1.ArgocdOIDC{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					GithubConnector: &v1beta1.ArgocdGithubConnector{
						Config: &v1beta1.ArgocdGithubConfig{
							ClientID:     &secret.Secret{Masterkey: masterkey},
							ClientSecret: &secret.Secret{Masterkey: masterkey},
						},
					},
					GitlabConnector: &v1beta1.ArgocdGitlabConnector{
						Config: &v1beta1.ArgocdGitlabConfig{
							ClientID:     &secret.Secret{Masterkey: masterkey},
							ClientSecret: &secret.Secret{Masterkey: masterkey},
						},
					},
					GoogleConnector: &v1beta1.ArgocdGoogleConnector{
						Config: &v1beta1.ArgocdGoogleConfig{
							ClientID:           &secret.Secret{Masterkey: masterkey},
							ClientSecret:       &secret.Secret{Masterkey: masterkey},
							ServiceAccountJSON: &secret.Secret{Masterkey: masterkey},
						},
					},
				},
				//Repositories: []*v1beta1.ArgocdRepository{{
				//	Username:    &secret.Secret{Masterkey: masterkey},
				//	Password:    &secret.Secret{Masterkey: masterkey},
				//	Certificate: &secret.Secret{Masterkey: masterkey},
				//}},
				//CustomImage: &v1beta1.ArgocdCustomImage{
				//	GopassStores: []*v1beta1.ArgocdGopassStore{{
				//		SSHKey: &secret.Secret{Masterkey: masterkey},
				//		GPGKey: &secret.Secret{Masterkey: masterkey},
				//	},
				//	},
				//},
			},
		},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}

func SecretFunc(orb *orbconfig.Orb) secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := parseToolset(desiredTree, orb.Masterkey)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		return getSecretsMap(desiredKind), nil
	}
}

func getSecretsMap(desiredKind *v1beta1.Toolset) map[string]*secret.Secret {
	return map[string]*secret.Secret{
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
}
