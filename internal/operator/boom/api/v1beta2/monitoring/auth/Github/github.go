package github

import (
	"github.com/caos/orbos/internal/secret"
)

type Auth struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	AllowedOrganizations       []string         `json:"allowedOrganizations,omitempty" yaml:"allowedOrganizations,omitempty"`
	TeamIDs                    []string         `json:"teamIDs,omitempty" yaml:"teamIDs,omitempty"`
}
