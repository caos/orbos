package auth

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/github"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/gitlab"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/google"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/oidc"
)

type Auth struct {
	OIDC            *oidc.OIDC        `json:"oidc,omitempty" yaml:"oidc,omitempty"`
	GithubConnector *github.Connector `json:"github,omitempty" yaml:"github,omitempty"`
	GitlabConnector *gitlab.Connector `json:"gitlab,omitempty" yaml:"gitlab,omitempty"`
	GoogleConnector *google.Connector `json:"google,omitempty" yaml:"google,omitempty"`
}
