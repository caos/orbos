package auth

import (
	"strings"

	"github.com/caos/orbos/v5/pkg/secret/read"

	gitlab "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Gitlab"
)

func GetGitlabAuthConfig(spec *gitlab.Auth) (map[string]string, error) {
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

	allowedGroups := strings.Join(spec.AllowedGroups, " ")

	return map[string]string{
		"enabled":        "true",
		"allow_sign_up":  "false",
		"client_id":      clientID,
		"client_secret":  clientSecret,
		"scopes":         "api",
		"auth_url":       "https://gitlab.com/oauth/authorize",
		"token_url":      "https://gitlab.com/oauth/token",
		"api_url":        "https://gitlab.com/api/v4",
		"allowed_groups": allowedGroups,
	}, nil
}
