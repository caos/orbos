package auth

import (
	"strings"

	"github.com/caos/orbos/v5/pkg/secret/read"

	generic "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Generic"
)

func GetGenericOAuthConfig(spec *generic.Auth) (map[string]string, error) {
	clientID, err := read.GetSecretValueOnlyIncluster(spec.ClientID, spec.ExistingClientIDSecret)
	if err != nil {
		return nil, err
	}

	clientSecret, err := read.GetSecretValueOnlyIncluster(spec.ClientSecret, spec.ExistingClientSecretSecret)
	if err != nil {
		return nil, err
	}

	if clientID == "" || clientSecret == "" {
		return nil, nil
	}

	allowedDomains := strings.Join(spec.AllowedDomains, " ")
	scopes := strings.Join(spec.Scopes, " ")

	return map[string]string{
		"enabled":         "true",
		"allow_sign_up":   "true",
		"client_id":       clientID,
		"client_secret":   clientSecret,
		"scopes":          scopes,
		"auth_url":        spec.AuthURL,
		"token_url":       spec.TokenURL,
		"api_url":         spec.APIURL,
		"allowed_domains": allowedDomains,
	}, nil
}
