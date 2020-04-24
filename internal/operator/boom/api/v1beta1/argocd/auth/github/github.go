package github

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
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	Orgs                       []*Org           `json:"orgs,omitempty" yaml:"orgs,omitempty"`
	LoadAllGroups              bool             `json:"loadAllGroups,omitempty" yaml:"loadAllGroups,omitempty"`
	TeamNameField              string           `json:"teamNameField,omitempty" yaml:"teamNameField,omitempty"`
	UseLoginAsID               bool             `json:"useLoginAsID,omitempty" yaml:"useLoginAsID,omitempty"`
}

type Org struct {
	Name  string   `json:"name,omitempty" yaml:"name,omitempty"`
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
