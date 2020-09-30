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

type Config struct {
	ClientID *secret2.Secret `yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret2.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret2.Secret   `yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret2.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//List of hosted domains which are permitted to login
	HostedDomains []string `json:"hostedDomains,omitempty" yaml:"hostedDomains,omitempty"`
	//List of groups in the hosted domains which are permitted to login
	Groups             []string        `json:"groups,omitempty" yaml:"groups,omitempty"`
	ServiceAccountJSON *secret2.Secret `yaml:"serviceAccountJSON,omitempty"`
	//Existing secret with the JSON of the service account
	ExistingServiceAccountJSONSecret *secret2.Existing `json:"existingServiceAccountJSONSecret,omitempty" yaml:"existingServiceAccountJSONSecret,omitempty"`
	//File where the serviceAccountJSON will get persisted to impersonate G Suite admin
	ServiceAccountFilePath string `json:"serviceAccountFilePath,omitempty" yaml:"serviceAccountFilePath,omitempty"`
	//Email of a G Suite admin to impersonate
	AdminEmail string `json:"adminEmail,omitempty" yaml:"adminEmail,omitempty"`
}
