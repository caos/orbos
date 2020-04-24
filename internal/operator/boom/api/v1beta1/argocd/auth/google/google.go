package google

import (
	"github.com/caos/orbiter/internal/secret"
	"reflect"
)

type Connector struct {
	ID     string  `json:"id,omitempty" yaml:"id,omitempty"`
	Name   string  `json:"name,omitempty" yaml:"name,omitempty"`
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
	ClientID                         *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret           *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret                     *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret       *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	HostedDomains                    []string         `json:"hostedDomains,omitempty" yaml:"hostedDomains,omitempty"`
	Groups                           []string         `json:"groups,omitempty" yaml:"groups,omitempty"`
	ServiceAccountJSON               *secret.Secret   `yaml:"serviceAccountJSON,omitempty"`
	ExistingServiceAccountJSONSecret *secret.Existing `json:"existingServiceAccountJSONSecret,omitempty" yaml:"existingServiceAccountJSONSecret,omitempty"`
	ServiceAccountFilePath           string           `json:"serviceAccountFilePath,omitempty" yaml:"serviceAccountFilePath,omitempty"`
	AdminEmail                       string           `json:"adminEmail,omitempty" yaml:"adminEmail,omitempty"`
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
