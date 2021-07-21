package auth

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling/auth"
	"github.com/caos/orbos/pkg/secret/read"
)

type oidc struct {
	Name                   string            `yaml:"name,omitempty"`
	Issuer                 string            `yaml:"issuer,omitempty"`
	ClientID               string            `yaml:"clientID,omitempty"`
	ClientSecret           string            `yaml:"clientSecret,omitempty"`
	RequestedScopes        []string          `yaml:"requestedScopes,omitempty"`
	RequestedIDTokenClaims map[string]*Claim `yaml:"requestedIDTokenClaims,omitempty"`
}
type Claim struct {
	Essential bool     `yaml:"essential,omitempty"`
	Values    []string `yaml:"values,omitempty"`
}

func GetOIDC(spec *auth.Auth) (string, error) {
	if spec == nil || spec.OIDC == nil {
		return "", nil
	}

	clientID, err := read.GetSecretValueOnlyIncluster(spec.OIDC.ClientID, spec.OIDC.ExistingClientIDSecret)
	if err != nil {
		return "", err
	}

	clientSecret, err := read.GetSecretValueOnlyIncluster(spec.OIDC.ClientSecret, spec.OIDC.ExistingClientSecretSecret)
	if err != nil {
		return "", err
	}

	if clientID == "" || clientSecret == "" {
		return "", nil
	}

	var claims map[string]*Claim
	if len(spec.OIDC.RequestedIDTokenClaims) > 0 {
		claims = make(map[string]*Claim, 0)
		for k, v := range spec.OIDC.RequestedIDTokenClaims {
			claims[k] = &Claim{
				Essential: v.Essential,
				Values:    v.Values,
			}
		}
	}

	oidc := &oidc{
		Name:                   spec.OIDC.Name,
		Issuer:                 spec.OIDC.Issuer,
		ClientID:               clientID,
		ClientSecret:           clientSecret,
		RequestedScopes:        spec.OIDC.RequestedScopes,
		RequestedIDTokenClaims: claims,
	}

	data, err := yaml.Marshal(oidc)
	if err != nil {
		return "", fmt.Errorf("error while generating argocd oidc configuration: %w", err)
	}
	return string(data), nil
}
