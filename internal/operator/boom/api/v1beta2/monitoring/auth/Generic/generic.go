package generic

import (
	"github.com/caos/orbos/internal/secret"
)

type Auth struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	Scopes                     []string         `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	AuthURL                    string           `json:"authURL,omitempty" yaml:"authURL,omitempty"`
	TokenURL                   string           `json:"tokenURL,omitempty" yaml:"tokenURL,omitempty"`
	APIURL                     string           `json:"apiURL,omitempty" yaml:"apiURL,omitempty"`
	AllowedDomains             []string         `json:"allowedDomains,omitempty" yaml:"allowedDomains,omitempty"`
}
