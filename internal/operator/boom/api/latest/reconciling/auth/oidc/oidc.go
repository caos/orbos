package oidc

import (
	secret2 "github.com/caos/orbos/pkg/secret"
)

type OIDC struct {
	//Internal name of the OIDC provider
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Issuer of the OIDC provider
	Issuer   string          `json:"issuer,omitempty" yaml:"issuer,omitempty"`
	ClientID *secret2.Secret `yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret2.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret2.Secret   `yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret2.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//Optional set of OIDC scopes to request. If omitted, defaults to: ["openid", "profile", "email", "groups"]
	RequestedScopes []string `json:"requestedScopes,omitempty" yaml:"requestedScopes,omitempty"`
	//Optional set of OIDC claims to request on the ID token.
	RequestedIDTokenClaims map[string]Claim `json:"requestedIDTokenClaims,omitempty" yaml:"requestedIDTokenClaims,omitempty"`
}

func (c *OIDC) IsZero() bool {
	if (c.ClientID == nil || c.ClientID.IsZero()) &&
		(c.ClientSecret == nil || c.ClientSecret.IsZero()) &&
		c.ExistingClientIDSecret == nil &&
		c.ExistingClientSecretSecret == nil &&
		c.Name == "" &&
		c.Issuer == "" &&
		c.RequestedScopes == nil &&
		c.RequestedIDTokenClaims == nil {
		return true
	}
	return false
}

type Claim struct {
	//Define if the claim is required, otherwise the login will fail
	Essential bool `json:"essential,omitempty" yaml:"essential,omitempty"`
	//Required values of the claim, otherwise hte login will fail
	Values []string `json:"values,omitempty" yaml:"values,omitempty"`
}
