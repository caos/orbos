package auth

import (
	generic "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Generic"
	helper2 "github.com/caos/orbos/internal/utils/helper"
	"strings"
)

func GetGenericOAuthConfig(spec *generic.Auth) (map[string]string, error) {
	clientID, err := helper2.GetSecretValue(spec.ClientID, spec.ExistingClientIDSecret)
	if err != nil {
		return nil, err
	}

	clientSecret, err := helper2.GetSecretValue(spec.ClientSecret, spec.ExistingClientSecretSecret)
	if err != nil {
		return nil, err
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
