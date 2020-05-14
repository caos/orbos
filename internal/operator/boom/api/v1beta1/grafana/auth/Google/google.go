package google

import (
	"github.com/caos/orbos/internal/secret"
	"reflect"
)

type Auth struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	AllowedDomains             []string         `json:"allowedDomains,omitempty" yaml:"allowedDomains,omitempty"`
}

func ClearEmpty(x *Auth) *Auth {
	if x == nil {
		return nil
	}

	marshaled := Auth{
		ClientID:                   secret.ClearEmpty(x.ClientID),
		ExistingClientIDSecret:     x.ExistingClientIDSecret,
		ClientSecret:               secret.ClearEmpty(x.ClientSecret),
		ExistingClientSecretSecret: x.ExistingClientSecretSecret,
		AllowedDomains:             x.AllowedDomains,
	}
	if reflect.DeepEqual(marshaled, Auth{}) {
		return nil
	}
	return &marshaled
}

func (g *Auth) MarshalYAML() (interface{}, error) {
	type Alias Auth
	return &Alias{
		ClientID:                   secret.ClearEmpty(g.ClientID),
		ExistingClientIDSecret:     g.ExistingClientIDSecret,
		ClientSecret:               secret.ClearEmpty(g.ClientSecret),
		ExistingClientSecretSecret: g.ExistingClientSecretSecret,
		AllowedDomains:             g.AllowedDomains,
	}, nil
}
