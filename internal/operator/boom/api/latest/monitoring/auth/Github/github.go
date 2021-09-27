package github

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
	// Organizations allowed to login
	AllowedOrganizations []string `json:"allowedOrganizations,omitempty" yaml:"allowedOrganizations,omitempty"`
	// TeamIDs where the user is required to have at least one membership
	TeamIDs []string `json:"teamIDs,omitempty" yaml:"teamIDs,omitempty"`
}

func (a *Auth) IsZero() bool {
	if (a.ClientID == nil || a.ClientID.IsZero()) &&
		(a.ClientSecret == nil || a.ClientSecret.IsZero()) &&
		a.ExistingClientIDSecret == nil &&
		a.ExistingClientSecretSecret == nil &&
		a.AllowedOrganizations == nil &&
		a.TeamIDs == nil {
		return true
	}

	return false
}
