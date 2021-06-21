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
	//Name of the user used for all actions
	User         *secret.Secret   `json:"user,omitempty" yaml:"user,omitempty"`
	ExistingUser *secret.Existing `json:"existinguser,omitempty" yaml:"existinguser,omitempty"`
	//API-key used for all actions besides the "Origin CA"-certificates
	APIKey         *secret.Secret   `json:"apikey,omitempty" yaml:"apikey,omitempty"`
	ExistingAPIKey *secret.Existing `json:"existingapikey,omitempty" yaml:"existingapikey,omitempty"`
	//User service key used for maintaining the "Origin CA"-certificates
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
	//Internal name of the group used in the firewall rules
	Name string `yaml:"name"`
	//List of strings which could contain for example IPs
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
	//Name of the subdomain
	Subdomain string `yaml:"subdomain"`
	//IP which is the target of the DNS entry
	IP string `yaml:"ip"`
	//Flag if DNS entry is proxied by cloudflare
	Proxied bool `yaml:"proxied"`
	//Time-to-live for the DNS entry
	TTL int `yaml:"ttl"`
	//Type of the DNS entry
	Type string `yaml:"type"`
	//The priority of the rule to allow control of processing order. A lower number indicates high priority. If not provided, any rules with a priority will be sequenced before those witho
	Priority int `yaml:"priority"`
}

type Rule struct {
	// Description given to the firewall rule
	Description string `yaml:"description"`
	//The priority of the rule to allow control of processing order. A lower number indicates high priority. If not provided, any rules with a priority will be sequenced before those without.
	Priority int `yaml:"priority"`
	//The action to apply to a matched request. Note that action "log" is only available for enterprise customers.
	Action string `yaml:"action"`
	//Definition of the filter used
	Filters []*Filter `yaml:"filters"`
}

type Filter struct {
	//A note that you can use to describe the purpose of the filter
	Description string `yaml:"description"`
	//List of targets
	Targets []string `yaml:"targets"`
	//List of target groups defined in the group attribute
	TargetGroups []string `yaml:"targetgroups"`
	//List of sources
	Sources []string `yaml:"sources"`
	//List of source groups defined in the group attribute
	SourceGroups []string `yaml:"sourcegroups"`
	//List of targets used in a "contains" expression
	ContainsTargets []string `yaml:"containstargets"`
	//List of target groups used in a "contains" expression
	ContainsTargetsGroups []string `yaml:"containstargetsgroups"`
	//Flag if SSL is required or not
	SSL string `yaml:"ssl"`
}
