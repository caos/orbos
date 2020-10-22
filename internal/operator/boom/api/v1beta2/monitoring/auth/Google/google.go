package google

import (
	"github.com/caos/orbos/internal/secret"
)

type Auth struct {
	ClientID *secret.Secret `yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret.Secret   `yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
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
