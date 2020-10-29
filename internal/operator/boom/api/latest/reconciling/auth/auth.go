package auth

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling/auth/github"
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling/auth/gitlab"
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling/auth/google"
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling/auth/oidc"
	"github.com/caos/orbos/internal/secret"
)

type Auth struct {
	//Configuration for SSO with a generic OIDC provider
	OIDC *oidc.OIDC `json:"oidc,omitempty" yaml:"oidc,omitempty"`
	//Configuration for SSO with Github
	GithubConnector *github.Connector `json:"github,omitempty" yaml:"github,omitempty"`
	//Configuration for SSO with Gitlab
	GitlabConnector *gitlab.Connector `json:"gitlab,omitempty" yaml:"gitlab,omitempty"`
	//Configuration for SSO with Google
	GoogleConnector *google.Connector `json:"google,omitempty" yaml:"google,omitempty"`
}

func (a *Auth) IsZero() bool {
	if (a.OIDC == nil || a.OIDC.IsZero()) &&
		(a.GithubConnector == nil || a.GithubConnector.IsZero()) &&
		(a.GitlabConnector == nil || a.GitlabConnector.IsZero()) &&
		(a.GoogleConnector == nil || a.GoogleConnector.IsZero()) {
		return true
	}
	return false
}

func (a *Auth) InitSecrets() {
	if a.OIDC == nil {
		a.OIDC = &oidc.OIDC{}
	}
	if a.OIDC.ClientID == nil {
		a.OIDC.ClientID = &secret.Secret{}
	}
	if a.OIDC.ClientSecret == nil {
		a.OIDC.ClientSecret = &secret.Secret{}
	}

	if a.GoogleConnector == nil {
		a.GoogleConnector = &google.Connector{}
	}
	if a.GoogleConnector.Config == nil {
		a.GoogleConnector.Config = &google.Config{}
	}
	if a.GoogleConnector.Config.ClientID == nil {
		a.GoogleConnector.Config.ClientID = &secret.Secret{}
	}
	if a.GoogleConnector.Config.ClientSecret == nil {
		a.GoogleConnector.Config.ClientSecret = &secret.Secret{}
	}
	if a.GoogleConnector.Config.ServiceAccountJSON == nil {
		a.GoogleConnector.Config.ServiceAccountJSON = &secret.Secret{}
	}

	if a.GithubConnector == nil {
		a.GithubConnector = &github.Connector{}
	}
	if a.GithubConnector.Config == nil {
		a.GithubConnector.Config = &github.Config{}
	}
	if a.GithubConnector.Config.ClientID == nil {
		a.GithubConnector.Config.ClientID = &secret.Secret{}
	}
	if a.GithubConnector.Config.ClientSecret == nil {
		a.GithubConnector.Config.ClientSecret = &secret.Secret{}
	}

	if a.GitlabConnector == nil {
		a.GitlabConnector = &gitlab.Connector{}
	}
	if a.GitlabConnector.Config == nil {
		a.GitlabConnector.Config = &gitlab.Config{}
	}
	if a.GitlabConnector.Config.ClientID == nil {
		a.GitlabConnector.Config.ClientID = &secret.Secret{}
	}
	if a.GitlabConnector.Config.ClientSecret == nil {
		a.GitlabConnector.Config.ClientSecret = &secret.Secret{}
	}
}
