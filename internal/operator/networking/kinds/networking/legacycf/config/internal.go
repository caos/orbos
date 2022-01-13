package config

import (
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"
)

type InternalConfig struct {
	Domains            []*InternalDomain `yaml:"domains"`
	Groups             []*Group          `yaml:"groups"`
	Credentials        *Credentials
	Prefix             string
	Namespace          string
	OriginCASecretName string
	Labels             *labels.API
}

type Credentials struct {
	User                   *secret.Secret   `json:"user,omitempty" yaml:"user,omitempty"`
	ExistingUser           *secret.Existing `json:"existinguser,omitempty" yaml:"existinguser,omitempty"`
	APIKey                 *secret.Secret   `json:"apikey,omitempty" yaml:"apikey,omitempty"`
	ExistingAPIKey         *secret.Existing `json:"existingapikey,omitempty" yaml:"existingapikey,omitempty"`
	UserServiceKey         *secret.Secret   `json:"userservicekey,omitempty" yaml:"userservicekey,omitempty"`
	ExistingUserServiceKey *secret.Existing `json:"existinguserservicekey,omitempty" yaml:"existinguserservicekey,omitempty"`
}

func (c *Credentials) IsZero() bool {
	if ((c.User == nil || c.User.IsZero()) && (c.ExistingUser == nil || c.ExistingUser.IsZero())) &&
		((c.APIKey == nil || c.APIKey.IsZero()) && (c.ExistingAPIKey == nil || c.ExistingAPIKey.IsZero())) &&
		((c.UserServiceKey == nil || c.UserServiceKey.IsZero()) && (c.ExistingUserServiceKey == nil || c.ExistingUserServiceKey.IsZero())) {
		return true
	}
	return false
}

type Group struct {
	Name string   `yaml:"name"`
	List []string `yaml:"list"`
}

type InternalDomain struct {
	Domain     string       `yaml:"domain"`
	Origin     *Origin      `yaml:"origin"`
	Subdomains []*Subdomain `yaml:"subdomains"`
	Rules      []*Rule      `yaml:"rules"`
}

type Origin struct {
	Key         *secret.Secret
	Certificate *secret.Secret
}

type Subdomain struct {
	Subdomain string  `yaml:"subdomain"`
	IP        string  `yaml:"ip"`
	Proxied   *bool   `yaml:"proxied"`
	TTL       int     `yaml:"ttl"`
	Type      string  `yaml:"type"`
	Priority  *uint16 `yaml:"priority"`
}

type Rule struct {
	Description string    `yaml:"description"`
	Priority    int       `yaml:"priority"`
	Action      string    `yaml:"action"`
	Filters     []*Filter `yaml:"filters"`
}

type Filter struct {
	Description           string   `yaml:"description"`
	Targets               []string `yaml:"targets"`
	TargetGroups          []string `yaml:"targetgroups"`
	Sources               []string `yaml:"sources"`
	SourceGroups          []string `yaml:"sourcegroups"`
	ContainsTargets       []string `yaml:"containstargets"`
	ContainsTargetsGroups []string `yaml:"containstargetsgroups"`
	SSL                   string   `yaml:"ssl"`
}
