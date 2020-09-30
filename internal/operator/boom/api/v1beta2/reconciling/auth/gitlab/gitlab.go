package gitlab

import (
	secret2 "github.com/caos/orbos/pkg/secret"
)

type Connector struct {
	//Internal id of the gitlab provider
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	//Internal name of the gitlab provider
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Configuration for the gitlab provider
	Config *Config `json:"config,omitempty" yaml:"config,omitempty"`
}

type Config struct {
	ClientID *secret2.Secret `yaml:"clientID,omitempty"`
	//Existing secret with the clientID
	ExistingClientIDSecret *secret2.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret           *secret2.Secret   `yaml:"clientSecret,omitempty"`
	//Existing secret with the clientSecret
	ExistingClientSecretSecret *secret2.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	//BaseURL of the gitlab instance
	BaseURL string `json:"baseURL,omitempty" yaml:"baseURL,omitempty"`
	//Optional groups whitelist, communicated through the "groups" scope
	Groups []string `json:"groups,omitempty" yaml:"groups,omitempty"`
	//Flag which will switch from using the internal GitLab id to the users handle (@mention) as the user id
	UseLoginAsID bool `json:"useLoginAsID,omitempty" yaml:"useLoginAsID,omitempty"`
}
