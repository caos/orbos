package github

import (
	"github.com/caos/orbos/internal/secret"
	"reflect"
)

type Connector struct {
	//Internal id of the github provider
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	//Internal name of the github provider
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Configuration for the github provider
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
	//Required membership to organization in github
	Orgs []*Org `json:"orgs,omitempty" yaml:"orgs,omitempty"`
	//Flag which indicates that all user groups and teams should be loaded
	LoadAllGroups bool `json:"loadAllGroups,omitempty" yaml:"loadAllGroups,omitempty"`
	//Optional choice between 'name' (default), 'slug', or 'both'
	TeamNameField string `json:"teamNameField,omitempty" yaml:"teamNameField,omitempty"`
	//Flag which will switch from using the internal GitHub id to the users handle (@mention) as the user id
	UseLoginAsID bool `json:"useLoginAsID,omitempty" yaml:"useLoginAsID,omitempty"`
}

type Org struct {
	//Name of the organization
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Name of the required team in the organization
	Teams []string `json:"teams,omitempty" yaml:"teams,omitempty"`
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
		Orgs:                       x.Orgs,
		LoadAllGroups:              x.LoadAllGroups,
		TeamNameField:              x.TeamNameField,
		UseLoginAsID:               x.UseLoginAsID,
	}

	if reflect.DeepEqual(marshaled, Config{}) {
		return nil
	}
	return &marshaled
}
