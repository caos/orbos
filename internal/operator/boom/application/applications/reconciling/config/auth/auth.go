package auth

import (
	"fmt"
	"strings"

	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling"
	"github.com/caos/orbos/mntr"
)

type Connectors struct {
	Connectors []*connector `yaml:"connectors,omitempty"`
}

type connector struct {
	Type   string
	Name   string
	ID     string
	Config interface{}
}

func GetDexConfigFromSpec(monitor mntr.Monitor, spec *reconciling.Reconciling) *Connectors {
	logFields := map[string]interface{}{
		"application": "argocd",
	}

	connectors := make([]*connector, 0)

	if spec.Auth == nil ||
		((spec.Auth.OIDC == nil || (spec.Auth.OIDC.ClientSecret == nil || spec.Auth.OIDC.ClientSecret.Value == "") && (spec.Auth.OIDC.ExistingClientSecretSecret == nil || spec.Auth.OIDC.ExistingClientSecretSecret.Name == "")) &&
			(spec.Auth.GithubConnector == nil || (spec.Auth.GithubConnector.Config.ClientSecret == nil || spec.Auth.GithubConnector.Config.ClientSecret.Value == "") && (spec.Auth.GithubConnector.Config.ExistingClientSecretSecret == nil || spec.Auth.GithubConnector.Config.ExistingClientSecretSecret.Name == "")) &&
			(spec.Auth.GitlabConnector == nil || (spec.Auth.GitlabConnector.Config.ClientSecret == nil || spec.Auth.GitlabConnector.Config.ClientSecret.Value == "") && (spec.Auth.GitlabConnector.Config.ExistingClientSecretSecret == nil || spec.Auth.GitlabConnector.Config.ExistingClientSecretSecret.Name == "")) &&
			(spec.Auth.GoogleConnector == nil || (spec.Auth.GoogleConnector.Config.ClientSecret == nil || spec.Auth.GoogleConnector.Config.ClientSecret.Value == "") && (spec.Auth.GoogleConnector.Config.ExistingClientSecretSecret == nil || spec.Auth.GoogleConnector.Config.ExistingClientSecretSecret.Name == ""))) {
		return &Connectors{Connectors: connectors}
	}

	if spec.Network == nil || spec.Network.Domain == "" {
		monitor.WithFields(logFields).Info("No auth connectors configured as no rootUrl is defined")
		return &Connectors{Connectors: connectors}
	}
	redirect := strings.Join([]string{"https://", spec.Network.Domain, "/api/dex/callback"}, "")

	if spec.Auth.GithubConnector != nil {
		github, err := getGithub(spec.Auth.GithubConnector, redirect)
		if err == nil && github != nil {
			connectors = append(connectors, &connector{
				Name:   spec.Auth.GithubConnector.Name,
				ID:     spec.Auth.GithubConnector.ID,
				Type:   "github",
				Config: github,
			})
		} else {
			monitor.WithFields(logFields).Error(fmt.Errorf("error while creating configuration for github connector: %w", err))
		}
	}

	if spec.Auth.GitlabConnector != nil {
		gitlab, err := getGitlab(spec.Auth.GitlabConnector, redirect)
		if err == nil && gitlab != nil {
			connectors = append(connectors, &connector{
				Name:   spec.Auth.GitlabConnector.Name,
				ID:     spec.Auth.GitlabConnector.ID,
				Type:   "gitlab",
				Config: gitlab,
			})
		} else {
			monitor.WithFields(logFields).Error(fmt.Errorf("error while creating configuration for gitlab connector: %w", err))
		}
	}

	if spec.Auth.GoogleConnector != nil {
		google, err := getGoogle(spec.Auth.GoogleConnector, redirect)
		if err == nil && google != nil {
			connectors = append(connectors, &connector{
				Name:   spec.Auth.GoogleConnector.Name,
				ID:     spec.Auth.GoogleConnector.ID,
				Type:   "oidc",
				Config: google,
			})
		} else {
			monitor.WithFields(logFields).Error(fmt.Errorf("error while creating configuration for google connector: %w", err))
		}
	}

	if len(connectors) > 0 {
		logFields["connectors"] = len(connectors)
		monitor.WithFields(logFields).Debug("Created dex configuration")
		return &Connectors{Connectors: connectors}
	}
	return nil
}
