package github

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
	Orgs                       []*Org           `json:"orgs,omitempty" yaml:"orgs,omitempty"`
	LoadAllGroups              bool             `json:"loadAllGroups,omitempty" yaml:"loadAllGroups,omitempty"`
	TeamNameField              string           `json:"teamNameField,omitempty" yaml:"teamNameField,omitempty"`
	UseLoginAsID               bool             `json:"useLoginAsID,omitempty" yaml:"useLoginAsID,omitempty"`
}

type Org struct {
	Name  string   `json:"name,omitempty" yaml:"name,omitempty"`
	Teams []string `json:"teams,omitempty" yaml:"teams,omitempty"`
}
