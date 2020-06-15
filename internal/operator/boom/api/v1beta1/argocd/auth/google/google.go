package google

import (
	"github.com/caos/orbos/internal/secret"
	"reflect"
)

type Connector struct {
	//Internal id of the google provider
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	//Internal name of the google provider
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Configuration for the google provider
	Config *Config `json:"config,omitempty" yaml:"config,omitempty"`
}

func ClearEmpty(x *Connector) *Connector {
	if x == nil {
		return nil
	}

	marshaled := Connector{
		ID:     x.ID,
		Name:   x.Name,
		Config: ClearEmptyConfig(x.Config),
	}

	if reflect.DeepEqual(marshaled, Connector{}) {
		return nil
	}
	return &marshaled
}

type Config struct {
	ClientID *secret.Secret `yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret.Secret   `yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//List of hosted domains which are permitted to login
	HostedDomains []string `json:"hostedDomains,omitempty" yaml:"hostedDomains,omitempty"`
	//List of groups in the hosted domains which are permitted to login
	Groups             []string       `json:"groups,omitempty" yaml:"groups,omitempty"`
	ServiceAccountJSON *secret.Secret `yaml:"serviceAccountJSON,omitempty"`
	//Existing secret with the JSON of the service account
	ExistingServiceAccountJSONSecret *secret.Existing `json:"existingServiceAccountJSONSecret,omitempty" yaml:"existingServiceAccountJSONSecret,omitempty"`
	//File where the serviceAccountJSON will get persisted to impersonate G Suite admin
	ServiceAccountFilePath string `json:"serviceAccountFilePath,omitempty" yaml:"serviceAccountFilePath,omitempty"`
	//Email of a G Suite admin to impersonate
	AdminEmail string `json:"adminEmail,omitempty" yaml:"adminEmail,omitempty"`
}

func ClearEmptyConfig(x *Config) *Config {
	if x == nil {
		return nil
	}

	marshaled := Config{
		ClientID:                         secret.ClearEmpty(x.ClientID),
		ExistingClientIDSecret:           x.ExistingClientIDSecret,
		ClientSecret:                     secret.ClearEmpty(x.ClientSecret),
		ExistingClientSecretSecret:       x.ExistingClientSecretSecret,
		HostedDomains:                    x.HostedDomains,
		Groups:                           x.Groups,
		ServiceAccountJSON:               secret.ClearEmpty(x.ServiceAccountJSON),
		ExistingServiceAccountJSONSecret: x.ExistingServiceAccountJSONSecret,
		ServiceAccountFilePath:           x.ServiceAccountFilePath,
		AdminEmail:                       x.AdminEmail,
	}

	if reflect.DeepEqual(marshaled, Config{}) {
		return nil
	}
	return &marshaled
}
