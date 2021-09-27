package auth

import (
	generic "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Generic"
	github "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Github"
	gitlab "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Gitlab"
	google "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Google"
	"github.com/caos/orbos/v5/pkg/secret"
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
	if a.GenericOAuth.ExistingClientIDSecret == nil {
		a.GenericOAuth.ExistingClientIDSecret = &secret.Existing{}
	}
	if a.GenericOAuth.ExistingClientSecretSecret == nil {
		a.GenericOAuth.ExistingClientSecretSecret = &secret.Existing{}
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
	if a.Google.ExistingClientIDSecret == nil {
		a.Google.ExistingClientIDSecret = &secret.Existing{}
	}
	if a.Google.ExistingClientSecretSecret == nil {
		a.Google.ExistingClientSecretSecret = &secret.Existing{}
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
	if a.Github.ExistingClientIDSecret == nil {
		a.Github.ExistingClientIDSecret = &secret.Existing{}
	}
	if a.Github.ExistingClientSecretSecret == nil {
		a.Github.ExistingClientSecretSecret = &secret.Existing{}
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
	if a.Gitlab.ExistingClientIDSecret == nil {
		a.Gitlab.ExistingClientIDSecret = &secret.Existing{}
	}
	if a.Gitlab.ExistingClientSecretSecret == nil {
		a.Gitlab.ExistingClientSecretSecret = &secret.Existing{}
	}
}
