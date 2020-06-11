package auth

import (
	generic "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Generic"
	github "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Github"
	gitlab "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Gitlab"
	google "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Google"
	"reflect"
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

func (a *Auth) MarshalYAML() (interface{}, error) {
	type Alias Auth
	alias := &Alias{
		Google:       google.ClearEmpty(a.Google),
		Gitlab:       gitlab.ClearEmpty(a.Gitlab),
		Github:       github.ClearEmpty(a.Github),
		GenericOAuth: generic.ClearEmpty(a.GenericOAuth),
	}
	return alias, nil
}

func ClearEmpty(x *Auth) *Auth {
	if x == nil {
		return nil
	}
	marshaled := Auth{
		Google:       google.ClearEmpty(x.Google),
		Gitlab:       gitlab.ClearEmpty(x.Gitlab),
		Github:       github.ClearEmpty(x.Github),
		GenericOAuth: generic.ClearEmpty(x.GenericOAuth),
	}
	if reflect.DeepEqual(marshaled, Auth{}) {
		return nil
	}
	return &marshaled
}
