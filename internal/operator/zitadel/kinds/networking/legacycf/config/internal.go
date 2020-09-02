package config

import (
	"github.com/caos/orbos/internal/secret"
)

type InternalConfig struct {
	Domains            []*InternalDomain `yaml:"domains"`
	Groups             []*Group          `yaml:"groups"`
	Credentials        *Credentials
	Prefix             string
	Namespace          string
	OriginCASecretName string
	Labels             map[string]string
}

type Credentials struct {
	User           *secret.Secret
	APIKey         *secret.Secret
	UserServiceKey *secret.Secret
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
	Subdomain string `yaml:"subdomain"`
	IP        string `yaml:"ip"`
	Proxied   bool   `yaml:"proxied"`
	TTL       int    `yaml:"ttl"`
	Type      string `yaml:"type"`
	Priority  int    `yaml:"priority"`
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
