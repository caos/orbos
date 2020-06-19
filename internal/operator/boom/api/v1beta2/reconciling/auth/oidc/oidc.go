package oidc

import (
	"github.com/caos/orbos/internal/secret"
)

type OIDC struct {
	//Internal name of the OIDC provider
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Issuer of the OIDC provider
	Issuer   string         `json:"issuer,omitempty" yaml:"issuer,omitempty"`
	ClientID *secret.Secret `yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret.Secret   `yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//Optional set of OIDC scopes to request. If omitted, defaults to: ["openid", "profile", "email", "groups"]
	RequestedScopes []string `json:"requestedScopes,omitempty" yaml:"requestedScopes,omitempty"`
	//Optional set of OIDC claims to request on the ID token.
	RequestedIDTokenClaims map[string]Claim `json:"requestedIDTokenClaims,omitempty" yaml:"requestedIDTokenClaims,omitempty"`
}

type Claim struct {
	//Define if the claim is required, otherwise the login will fail
	Essential bool `json:"essential,omitempty" yaml:"essential,omitempty"`
	//Required values of the claim, otherwise hte login will fail
	Values []string `json:"values,omitempty" yaml:"values,omitempty"`
}
