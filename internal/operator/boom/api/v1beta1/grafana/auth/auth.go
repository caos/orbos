package auth

import (
	generic "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Generic"
	github "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Github"
	gitlab "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Gitlab"
	google "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Google"
	"github.com/caos/orbos/pkg/secret"
)

type Auth struct {
	//Configuration for SSO with Google
	Google *google.Auth `json:"google,omitempty" yaml:"google,omitempty"`
	//Configuration for SSO with Github
	Github *github.Auth `json:"github,omitempty" yaml:"github,omitempty"`
	//Configuration for SSO with Gitlab
	Gitlab *gitlab.Auth `json:"gitlab,omitempty" yaml:"gitlab,omitempty"`
	//Configuration for SSO with an generic OAuth provider
	GenericOAuth *generic.Auth `json:"genericOAuth,omitempty" yaml:"genericOAuth,omitempty"`
}

func (a *Auth) IsZero() bool {
	if (a.GenericOAuth == nil || a.GenericOAuth.IsZero()) &&
		(a.Github == nil || a.Github.IsZero()) &&
		(a.Gitlab == nil || a.Gitlab.IsZero()) &&
		(a.Google == nil || a.Google.IsZero()) {
		return true
	}

	return false
}

func (a *Auth) InitSecrets() {
	if a.GenericOAuth == nil {
		a.GenericOAuth = &generic.Auth{}
	}
	if a.GenericOAuth.ClientID == nil {
		a.GenericOAuth.ClientID = &secret.Secret{}
	}
	if a.GenericOAuth.ClientSecret == nil {
		a.GenericOAuth.ClientSecret = &secret.Secret{}
	}

	if a.Google == nil {
		a.Google = &google.Auth{}
	}
	if a.Google.ClientID == nil {
		a.Google.ClientID = &secret.Secret{}
	}
	if a.Google.ClientSecret == nil {
		a.Google.ClientSecret = &secret.Secret{}
	}

	if a.Github == nil {
		a.Github = &github.Auth{}
	}
	if a.Github.ClientID == nil {
		a.Github.ClientID = &secret.Secret{}
	}
	if a.Github.ClientSecret == nil {
		a.Github.ClientSecret = &secret.Secret{}
	}

	if a.Gitlab == nil {
		a.Gitlab = &gitlab.Auth{}
	}
	if a.Gitlab.ClientID == nil {
		a.Gitlab.ClientID = &secret.Secret{}
	}
	if a.Gitlab.ClientSecret == nil {
		a.Gitlab.ClientSecret = &secret.Secret{}
	}
}
