package auth

import (
	"github.com/caos/orbos/pkg/helper"
	"strings"

	google "github.com/caos/orbos/internal/operator/boom/api/latest/monitoring/auth/Google"
)

func GetGoogleAuthConfig(spec *google.Auth) (map[string]string, error) {
	clientID, err := helper.GetSecretValueOnlyIncluster(spec.ClientID, spec.ExistingClientIDSecret)
	if err != nil {
		return nil, err
	}

	clientSecret, err := helper.GetSecretValueOnlyIncluster(spec.ClientSecret, spec.ExistingClientSecretSecret)
	if err != nil {
		return nil, err
	}

	if clientID == "" || clientSecret == "" {
		return nil, nil
	}

	domains := strings.Join(spec.AllowedDomains, " ")

	return map[string]string{
		"enabled":         "true",
		"client_id":       string(clientID),
		"client_secret":   string(clientSecret),
		"scopes":          "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email",
		"auth_url":        "https://accounts.google.com/o/oauth2/auth",
		"token_url":       "https://accounts.google.com/o/oauth2/token",
		"allowed_domains": domains,
		"allow_sign_up":   "true",
	}, nil
}
