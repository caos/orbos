package auth

import (
	gitlab "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Gitlab"
	helper2 "github.com/caos/orbos/internal/utils/helper"
	"strings"
)

func GetGitlabAuthConfig(spec *gitlab.Auth) (map[string]string, error) {
	clientID, err := helper2.GetSecretValue(spec.ClientID, spec.ExistingClientIDSecret)
	if err != nil {
		return nil, err
	}

	clientSecret, err := helper2.GetSecretValue(spec.ClientSecret, spec.ExistingClientSecretSecret)
	if err != nil {
		return nil, err
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
