package auth

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth/github"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth/gitlab"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth/google"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth/oidc"
	"reflect"
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

func ClearEmpty(x *Auth) *Auth {
	if x == nil {
		return nil
	}

	marshaled := Auth{
		OIDC:            oidc.ClearEmpty(x.OIDC),
		GithubConnector: github.ClearEmpty(x.GithubConnector),
		GitlabConnector: gitlab.ClearEmpty(x.GitlabConnector),
		GoogleConnector: google.ClearEmpty(x.GoogleConnector),
	}

	if reflect.DeepEqual(marshaled, Auth{}) {
		return nil
	}
	return &marshaled
}
