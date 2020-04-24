package generic

import (
	"github.com/caos/orbiter/internal/secret"
	"reflect"
)

type Auth struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	Scopes                     []string         `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	AuthURL                    string           `json:"authURL,omitempty" yaml:"authURL,omitempty"`
	TokenURL                   string           `json:"tokenURL,omitempty" yaml:"tokenURL,omitempty"`
	APIURL                     string           `json:"apiURL,omitempty" yaml:"apiURL,omitempty"`
	AllowedDomains             []string         `json:"allowedDomains,omitempty" yaml:"allowedDomains,omitempty"`
}

func ClearEmpty(g *Auth) *Auth {
	if g == nil {
		return nil
	}

	marshaled := Auth{
		ClientID:                   secret.ClearEmpty(g.ClientID),
		ExistingClientIDSecret:     g.ExistingClientIDSecret,
		ClientSecret:               secret.ClearEmpty(g.ClientSecret),
		ExistingClientSecretSecret: g.ExistingClientSecretSecret,
		Scopes:                     g.Scopes,
		AuthURL:                    g.AuthURL,
		TokenURL:                   g.TokenURL,
		APIURL:                     g.APIURL,
		AllowedDomains:             g.AllowedDomains,
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
		Scopes:                     g.Scopes,
		AuthURL:                    g.AuthURL,
		TokenURL:                   g.TokenURL,
		APIURL:                     g.APIURL,
		AllowedDomains:             g.AllowedDomains,
	}, nil
}
