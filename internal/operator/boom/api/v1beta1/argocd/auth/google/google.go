package google

import (
	secret2 "github.com/caos/orbos/pkg/secret"
)

type Connector struct {
	//Internal id of the google provider
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	//Internal name of the google provider
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Configuration for the google provider
	Config *Config `json:"config,omitempty" yaml:"config,omitempty"`
}

func (c *Connector) IsZero() bool {
	if c.ID == "" &&
		c.Name == "" &&
		(c.Config == nil || c.Config.IsZero()) {
		return true
	}

	return false
}

func (c *Config) IsZero() bool {
	if (c.ClientID == nil || c.ClientID.IsZero()) &&
		(c.ClientSecret == nil || c.ClientSecret.IsZero()) &&
		(c.ServiceAccountJSON == nil || c.ServiceAccountJSON.IsZero()) &&
		c.ExistingClientIDSecret == nil &&
		c.ExistingClientSecretSecret == nil &&
		c.ExistingServiceAccountJSONSecret == nil &&
		c.HostedDomains == nil &&
		c.Groups == nil &&
		c.ServiceAccountFilePath == "" &&
		c.AdminEmail == "" {
		return true
	}
	return false
}

type Config struct {
	ClientID *secret2.Secret `json:"clientID,omitempty" yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret2.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret2.Secret   `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret2.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//List of hosted domains which are permitted to login
	HostedDomains []string `json:"hostedDomains,omitempty" yaml:"hostedDomains,omitempty"`
	//List of groups in the hosted domains which are permitted to login
	Groups             []string        `json:"groups,omitempty" yaml:"groups,omitempty"`
	ServiceAccountJSON *secret2.Secret `json:"serviceAccountJSON,omitempty" yaml:"serviceAccountJSON,omitempty"`
	//Existing secret with the JSON of the service account
	ExistingServiceAccountJSONSecret *secret2.Existing `json:"existingServiceAccountJSONSecret,omitempty" yaml:"existingServiceAccountJSONSecret,omitempty"`
	//File where the serviceAccountJSON will get persisted to impersonate G Suite admin
	ServiceAccountFilePath string `json:"serviceAccountFilePath,omitempty" yaml:"serviceAccountFilePath,omitempty"`
	//Email of a G Suite admin to impersonate
	AdminEmail string `json:"adminEmail,omitempty" yaml:"adminEmail,omitempty"`
}
