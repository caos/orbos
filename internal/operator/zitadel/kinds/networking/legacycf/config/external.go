package config

import (
	"errors"
	"github.com/caos/orbos/internal/operator/zitadel"

	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"

	"github.com/caos/orbos/internal/operator/orbiter"
)

type ExternalConfig struct {
	Verbose     bool
	Domain      string
	IP          orbiter.IPAddress
	Rules       []*Rule
	Groups      []*Group     `yaml:"groups"`
	Credentials *Credentials `yaml:"credentials"`
	Prefix      string       `yaml:"prefix"`
}

func (e *ExternalConfig) Internal(namespace string, originCASecretName string, labels map[string]string) (*InternalConfig, *current) {
	dom, curr := e.internalDomain()
	return &InternalConfig{
		Domains:            []*IntenalDomain{dom},
		Groups:             e.Groups,
		Credentials:        e.Credentials,
		Prefix:             e.Prefix,
		Namespace:          namespace,
		OriginCASecretName: originCASecretName,
		Labels:             labels,
	}, curr
}

func (e *ExternalConfig) Validate() error {
	if e == nil {
		return errors.New("domain not found")
	}
	if e.Domain == "" {
		return errors.New("No domain configured")
	}
	return e.IP.Validate()
}

func (e *ExternalConfig) internalDomain() (*IntenalDomain, *current) {
	return &IntenalDomain{
			Domain: e.Domain,
			Subdomains: []*Subdomain{
				subdomain("accounts", e.IP),
				subdomain("api", e.IP),
				subdomain("console", e.IP),
				subdomain("issuer", e.IP),
			},
			Rules: e.Rules,
		},
		&current{
			domain:            e.Domain,
			issureSubdomain:   "issuer",
			consoleSubdomain:  "console",
			apiSubdomain:      "api",
			accountsSubdomain: "accounts",
		}
}

func subdomain(subdomain string, ip orbiter.IPAddress) *Subdomain {
	return &Subdomain{
		Subdomain: subdomain,
		IP:        string(ip),
		Proxied:   true,
		TTL:       0,
		Type:      "A",
	}
}

var _ core.NetworkingCurrent = (*current)(nil)

type current struct {
	domain            string `yaml:"-"`
	issureSubdomain   string `yaml:"-"`
	consoleSubdomain  string `yaml:"-"`
	apiSubdomain      string `yaml:"-"`
	accountsSubdomain string `yaml:"-"`
	ReadyCertificate  zitadel.EnsureFunc
}

func (c *current) GetDomain() string {
	return c.domain
}
func (c *current) GetIssuerSubDomain() string {
	return c.issureSubdomain
}
func (c *current) GetConsoleSubDomain() string {
	return c.consoleSubdomain
}
func (c *current) GetAPISubDomain() string {
	return c.apiSubdomain
}
func (c *current) GetAccountsSubDomain() string {
	return c.accountsSubdomain
}
func (c *current) GetReadyCertificate() zitadel.EnsureFunc {
	return c.ReadyCertificate
}
