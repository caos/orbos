package auth

import (
	generic "github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth/Generic"
	github "github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth/Github"
	gitlab "github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth/Gitlab"
	google "github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth/Google"
)

type Auth struct {
	Google       *google.Auth  `json:"google,omitempty" yaml:"google,omitempty"`
	Github       *github.Auth  `json:"github,omitempty" yaml:"github,omitempty"`
	Gitlab       *gitlab.Auth  `json:"gitlab,omitempty" yaml:"gitlab,omitempty"`
	GenericOAuth *generic.Auth `json:"genericOAuth,omitempty" yaml:"genericOAuth,omitempty"`
}
