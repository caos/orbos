package auth

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/github"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/gitlab"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/google"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/oidc"
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
