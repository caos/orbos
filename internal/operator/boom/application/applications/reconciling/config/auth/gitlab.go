package auth

import (
	"github.com/caos/orbos/v5/internal/operator/boom/api/latest/reconciling/auth/gitlab"
	"github.com/caos/orbos/v5/pkg/secret/read"
)

type gitlabConnector struct {
	ClientID     string   `yaml:"clientID,omitempty"`
	ClientSecret string   `yaml:"clientSecret,omitempty"`
	RedirectURI  string   `yaml:"redirectURI,omitempty"`
	BaseURL      string   `yaml:"baseURL,omitempty"`
	Groups       []string `yaml:"groups,omitempty"`
	UseLoginAsID bool     `yaml:"useLoginAsID,omitempty"`
}

func getGitlab(spec *gitlab.Connector, redirect string) (interface{}, error) {
	clientID, err := read.GetSecretValueOnlyIncluster(spec.Config.ClientID, spec.Config.ExistingClientIDSecret)
	if err != nil {
		return nil, err
	}

	clientSecret, err := read.GetSecretValueOnlyIncluster(spec.Config.ClientSecret, spec.Config.ExistingClientSecretSecret)
	if err != nil {
		return nil, err
	}

	if clientID == "" || clientSecret == "" {
		return nil, nil
	}

	var groups []string
	if len(spec.Config.Groups) > 0 {
		groups = make([]string, len(spec.Config.Groups))
		for k, v := range spec.Config.Groups {
			groups[k] = v
		}
	}

	gitlab := &gitlabConnector{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirect,
		Groups:       groups,
		UseLoginAsID: spec.Config.UseLoginAsID,
	}

	return gitlab, nil
}
