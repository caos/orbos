package oidc

import (
	"github.com/caos/orbos/internal/secret"
	"reflect"
)

type OIDC struct {
	Name                       string           `json:"name,omitempty" yaml:"name,omitempty"`
	Issuer                     string           `json:"issuer,omitempty" yaml:"issuer,omitempty"`
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	RequestedScopes            []string         `json:"requestedScopes,omitempty" yaml:"requestedScopes,omitempty"`
	RequestedIDTokenClaims     map[string]Claim `json:"requestedIDTokenClaims,omitempty" yaml:"requestedIDTokenClaims,omitempty"`
}

type Claim struct {
	Essential bool     `json:"essential,omitempty" yaml:"essential,omitempty"`
	Values    []string `json:"values,omitempty" yaml:"values,omitempty"`
}

func ClearEmpty(x *OIDC) *OIDC {
	if x == nil {
		return nil
	}

	marshaled := OIDC{
		Name:                       x.Name,
		Issuer:                     x.Issuer,
		ClientID:                   secret.ClearEmpty(x.ClientID),
		ExistingClientIDSecret:     x.ExistingClientIDSecret,
		ClientSecret:               secret.ClearEmpty(x.ClientSecret),
		ExistingClientSecretSecret: x.ExistingClientSecretSecret,
		RequestedScopes:            x.RequestedScopes,
		RequestedIDTokenClaims:     x.RequestedIDTokenClaims,
	}

	if reflect.DeepEqual(marshaled, OIDC{}) {
		return nil
	}
	return &marshaled
}
