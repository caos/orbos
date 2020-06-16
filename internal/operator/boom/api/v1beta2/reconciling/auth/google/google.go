package google

import (
	"github.com/caos/orbos/internal/secret"
)

type Connector struct {
	ID     string  `json:"id,omitempty" yaml:"id,omitempty"`
	Name   string  `json:"name,omitempty" yaml:"name,omitempty"`
	Config *Config `json:"config,omitempty" yaml:"config,omitempty"`
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
