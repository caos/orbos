package gitlab

import (
	"github.com/caos/orbos/internal/secret"
	"reflect"
)

type Connector struct {
	//Internal id of the gitlab provider
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	//Internal name of the gitlab provider
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Configuration for the gitlab provider
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
	//BaseURL of the gitlab instance
	BaseURL string `json:"baseURL,omitempty" yaml:"baseURL,omitempty"`
	//Optional groups whitelist, communicated through the "groups" scope
	Groups []string `json:"groups,omitempty" yaml:"groups,omitempty"`
	//Flag which will switch from using the internal GitLab id to the users handle (@mention) as the user id
	UseLoginAsID bool `json:"useLoginAsID,omitempty" yaml:"useLoginAsID,omitempty"`
}

func ClearEmptyConfig(x *Config) *Config {
	if x == nil {
		return nil
	}

	marshaled := Config{
		ClientID:                   secret.ClearEmpty(x.ClientID),
		ExistingClientIDSecret:     x.ExistingClientIDSecret,
		ClientSecret:               secret.ClearEmpty(x.ClientSecret),
		ExistingClientSecretSecret: x.ExistingClientSecretSecret,
		BaseURL:                    x.BaseURL,
		Groups:                     x.Groups,
		UseLoginAsID:               x.UseLoginAsID,
	}

	if reflect.DeepEqual(marshaled, Config{}) {
		return nil
	}
	return &marshaled
}
