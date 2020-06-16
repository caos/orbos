package gitlab

import (
	"github.com/caos/orbos/internal/secret"
)

type Connector struct {
	ID     string  `json:"id,omitempty" yaml:"id,omitempty"`
	Name   string  `json:"name,omitempty" yaml:"name,omitempty"`
	Config *Config `json:"config,omitempty" yaml:"config,omitempty"`
}

type Config struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	BaseURL                    string           `json:"baseURL,omitempty" yaml:"baseURL,omitempty"`
	Groups                     []string         `json:"groups,omitempty" yaml:"groups,omitempty"`
	UseLoginAsID               bool             `json:"useLoginAsID,omitempty" yaml:"useLoginAsID,omitempty"`
}
