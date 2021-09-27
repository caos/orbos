package grafana

import (
	"github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring"
	"github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/admin"
	"github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth"
	generic "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Generic"
	github "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Github"
	gitlab "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Gitlab"
	google "github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring/auth/Google"
	"github.com/caos/orbos/v5/internal/operator/boom/api/migrate/network"
	"github.com/caos/orbos/v5/internal/operator/boom/api/migrate/storage"
	"github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/grafana"
	"github.com/caos/orbos/v5/pkg/secret"
)

func V1beta1Tov1beta2(grafana *grafana.Grafana) *monitoring.Monitoring {
	newSpec := &monitoring.Monitoring{
		Deploy:       grafana.Deploy,
		NodeSelector: grafana.NodeSelector,
	}
	if grafana.Datasources != nil && len(grafana.Datasources) > 0 {
		datasources := make([]*monitoring.Datasource, 0)
		for _, v := range grafana.Datasources {
			newSpec.Datasources = append(newSpec.Datasources, &monitoring.Datasource{
				Name:      v.Name,
				Type:      v.Type,
				Url:       v.Url,
				Access:    v.Access,
				IsDefault: v.IsDefault,
			})
		}
		newSpec.Datasources = datasources
	}

	if grafana.DashboardProviders != nil && len(grafana.DashboardProviders) > 0 {
		providers := make([]*monitoring.Provider, 0)
		for _, v := range grafana.DashboardProviders {
			provider := &monitoring.Provider{
				ConfigMaps: nil,
				Folder:     v.Folder,
			}
			if v.ConfigMaps != nil && len(v.ConfigMaps) > 0 {
				for _, v := range v.ConfigMaps {
					provider.ConfigMaps = append(provider.ConfigMaps, v)
				}
			}
			newSpec.DashboardProviders = append(newSpec.DashboardProviders, provider)
		}
		newSpec.DashboardProviders = providers
	}

	if grafana.Admin != nil {
		newSpec.Admin = &admin.Admin{
			Username: grafana.Admin.Username,
			Password: grafana.Admin.Password,
			ExistingUsername: &secret.Existing{
				Name: grafana.Admin.ExistingSecret.Name,
				Key:  grafana.Admin.ExistingSecret.IDKey,
			},
			ExistingPassword: &secret.Existing{
				Name: grafana.Admin.ExistingSecret.Name,
				Key:  grafana.Admin.ExistingSecret.SecretKey,
			},
		}
	}

	if grafana.Network != nil {
		newSpec.Network = network.V1beta1Tov1beta2(grafana.Network)
	}

	if grafana.Storage != nil {
		newSpec.Storage = storage.V1beta1Tov1beta2(grafana.Storage)
	}

	if grafana.Auth != nil {
		oldAuth := grafana.Auth
		newAuth := auth.Auth{}
		if oldAuth.Google != nil {
			newAuth.Google = &google.Auth{
				ClientID:                   oldAuth.Google.ClientID,
				ExistingClientIDSecret:     oldAuth.Google.ExistingClientIDSecret,
				ClientSecret:               oldAuth.Google.ClientSecret,
				ExistingClientSecretSecret: oldAuth.Google.ExistingClientSecretSecret,
			}
			if oldAuth.Google.AllowedDomains != nil {
				domains := make([]string, 0)
				for _, v := range oldAuth.Google.AllowedDomains {
					domains = append(domains, v)
				}
				newAuth.Google.AllowedDomains = domains
			}
		}
		if oldAuth.Github != nil {
			newAuth.Github = &github.Auth{
				ClientID:                   oldAuth.Github.ClientID,
				ExistingClientIDSecret:     oldAuth.Github.ExistingClientIDSecret,
				ClientSecret:               oldAuth.Github.ClientSecret,
				ExistingClientSecretSecret: oldAuth.Github.ExistingClientSecretSecret,
			}
			if oldAuth.Github.AllowedOrganizations != nil {
				orgs := make([]string, 0)
				for _, v := range oldAuth.Github.AllowedOrganizations {
					orgs = append(orgs, v)
				}
				newAuth.Github.AllowedOrganizations = orgs
			}
			if oldAuth.Github.TeamIDs != nil {
				teams := make([]string, 0)
				for _, v := range oldAuth.Github.TeamIDs {
					teams = append(teams, v)
				}
				newAuth.Github.TeamIDs = teams
			}
		}

		if oldAuth.Gitlab != nil {
			newAuth.Gitlab = &gitlab.Auth{
				ClientID:                   oldAuth.Gitlab.ClientID,
				ExistingClientIDSecret:     oldAuth.Gitlab.ExistingClientIDSecret,
				ClientSecret:               oldAuth.Gitlab.ClientSecret,
				ExistingClientSecretSecret: oldAuth.Gitlab.ExistingClientSecretSecret,
			}
			if oldAuth.Gitlab.AllowedGroups != nil {
				groups := make([]string, 0)
				for _, v := range oldAuth.Gitlab.AllowedGroups {
					groups = append(groups, v)
				}
				newAuth.Gitlab.AllowedGroups = groups
			}
		}

		if oldAuth.GenericOAuth != nil {
			newAuth.GenericOAuth = &generic.Auth{
				ClientID:                   oldAuth.GenericOAuth.ClientID,
				ExistingClientIDSecret:     oldAuth.GenericOAuth.ExistingClientIDSecret,
				ClientSecret:               oldAuth.GenericOAuth.ClientSecret,
				ExistingClientSecretSecret: oldAuth.GenericOAuth.ExistingClientSecretSecret,
				AuthURL:                    oldAuth.GenericOAuth.AuthURL,
				TokenURL:                   oldAuth.GenericOAuth.TokenURL,
				APIURL:                     oldAuth.GenericOAuth.APIURL,
			}
			if oldAuth.GenericOAuth.AllowedDomains != nil {
				domains := make([]string, 0)
				for _, v := range oldAuth.GenericOAuth.AllowedDomains {
					domains = append(domains, v)
				}
				newAuth.GenericOAuth.AllowedDomains = domains
			}
			if oldAuth.GenericOAuth.Scopes != nil {
				scopes := make([]string, 0)
				for _, v := range oldAuth.GenericOAuth.Scopes {
					scopes = append(scopes, v)
				}
				newAuth.GenericOAuth.Scopes = scopes
			}
		}
		newSpec.Auth = &newAuth
	}

	return newSpec
}
