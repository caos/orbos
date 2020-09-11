package cloudflare

import (
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type Desired struct {
	Common *tree.Common `yaml:",inline"`
	Spec   *Spec
}

type Spec struct {
	Domains []*Domain `yaml:"domains"`
	Groups  []*Group  `yaml:"groups"`
}

type Group struct {
	Name string   `yaml:"name"`
	List []string `yaml:"list"`
}

type Domain struct {
	Domain     string       `yaml:"domain"`
	Subdomains []*Subdomain `yaml:"subdomains"`
	Rules      []*Rule      `yaml:"rules"`
}

type Subdomain struct {
	Subdomain string `yaml:"subdomain"`
	IP        string `yaml:"ip"`
	Proxied   bool   `yaml:"proxied"`
	TTL       int    `yaml:"ttl"`
	Type      string `yaml:"type"`
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

func parseDesired(desiredTree *tree.Tree) (*Desired, error) {
	desiredKind := &Desired{
		Common: desiredTree.Common,
		Spec:   &Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
