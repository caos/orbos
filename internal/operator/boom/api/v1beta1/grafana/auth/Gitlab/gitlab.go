package gitlab

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
	//Groups of gitlab allowed to login
	AllowedGroups []string `json:"allowedGroups,omitempty" yaml:"allowedGroups,omitempty"`
}
