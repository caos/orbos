package gitlab

import (
	secret2 "github.com/caos/orbos/pkg/secret"
)

type Auth struct {
	ClientID *secret2.Secret `yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret2.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret2.Secret   `yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret2.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//Groups of gitlab allowed to login
	AllowedGroups []string `json:"allowedGroups,omitempty" yaml:"allowedGroups,omitempty"`
}
