package google

import (
	secret2 "github.com/caos/orbos/v5/pkg/secret"
)

type Auth struct {
	ClientID *secret2.Secret `json:"clientID,omitempty" yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret2.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret2.Secret   `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret2.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//Domains allowed to login
	AllowedDomains []string `json:"allowedDomains,omitempty" yaml:"allowedDomains,omitempty"`
}

func (a *Auth) IsZero() bool {
	if (a.ClientID == nil || a.ClientID.IsZero()) &&
		(a.ClientSecret == nil || a.ClientSecret.IsZero()) &&
		a.ExistingClientIDSecret == nil &&
		a.ExistingClientSecretSecret == nil &&
		a.AllowedDomains == nil {
		return true
	}

	return false
}
