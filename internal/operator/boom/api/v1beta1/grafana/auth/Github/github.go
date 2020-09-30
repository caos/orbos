package github

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
	// Organizations allowed to login
	AllowedOrganizations []string `json:"allowedOrganizations,omitempty" yaml:"allowedOrganizations,omitempty"`
	// TeamIDs where the user is required to have at least one membership
	TeamIDs []string `json:"teamIDs,omitempty" yaml:"teamIDs,omitempty"`
}
