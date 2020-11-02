package argocd

import (
	"github.com/caos/orbos/internal/operator/boom/api/migrate/network"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/github"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/gitlab"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/google"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/oidc"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/repository"
)

func V1beta1Tov1beta2(old *argocd.Argocd) *reconciling.Reconciling {
	new := &reconciling.Reconciling{
		Deploy:       old.Deploy,
		NodeSelector: old.NodeSelector,
	}
	if old.Network != nil {
		new.Network = network.V1beta1Tov1beta2(old.Network)
	}
	if old.Rbac != nil {
		new.Rbac = &reconciling.Rbac{
			Csv:     old.Rbac.Csv,
			Default: old.Rbac.Default,
		}
		if old.Rbac.Scopes != nil {
			scopes := make([]string, 0)
			for _, v := range old.Rbac.Scopes {
				scopes = append(scopes, v)
			}
			new.Rbac.Scopes = scopes
		}
	}
	if old.CustomImage != nil {
		new.CustomImage = &reconciling.CustomImage{
			Enabled: old.CustomImage.Enabled,
		}
		if old.CustomImage.GopassStores != nil {
			stores := make([]*reconciling.GopassStore, 0)
			for _, v := range old.CustomImage.GopassStores {
				stores = append(stores, &reconciling.GopassStore{
					SSHKey:               v.SSHKey,
					ExistingSSHKeySecret: v.ExistingSSHKeySecret,
					GPGKey:               v.GPGKey,
					ExistingGPGKeySecret: v.ExistingGPGKeySecret,
					Directory:            v.Directory,
					StoreName:            v.StoreName,
				})
			}
			new.CustomImage.GopassStores = stores
		}
	}
	if old.Repositories != nil {
		repos := make([]*repository.Repository, 0)
		for _, v := range old.Repositories {
			repos = append(repos, &repository.Repository{
				Name:                      v.Name,
				URL:                       v.URL,
				Username:                  v.Username,
				ExistingUsernameSecret:    v.ExistingUsernameSecret,
				Password:                  v.Password,
				ExistingPasswordSecret:    v.ExistingPasswordSecret,
				Certificate:               v.Certificate,
				ExistingCertificateSecret: v.ExistingCertificateSecret,
			})
		}
		new.Repositories = repos
	}
	if old.Credentials != nil {
		creds := make([]*repository.Repository, 0)
		for _, v := range old.Credentials {
			creds = append(creds, &repository.Repository{
				Name:                      v.Name,
				URL:                       v.URL,
				Username:                  v.Username,
				ExistingUsernameSecret:    v.ExistingUsernameSecret,
				Password:                  v.Password,
				ExistingPasswordSecret:    v.ExistingPasswordSecret,
				Certificate:               v.Certificate,
				ExistingCertificateSecret: v.ExistingCertificateSecret,
			})
		}
		new.Credentials = creds
	}
	if old.KnownHosts != nil {
		hosts := make([]string, 0)
		for _, v := range old.KnownHosts {
			hosts = append(hosts, v)
		}
		new.KnownHosts = hosts
	}

	if old.Auth != nil {
		oldAuth := old.Auth
		newAuth := auth.Auth{}
		if oldAuth.GoogleConnector != nil {
			conn := &google.Connector{
				ID:   oldAuth.GoogleConnector.ID,
				Name: oldAuth.GoogleConnector.Name,
			}
			if oldAuth.GoogleConnector.Config != nil {
				oldConf := oldAuth.GoogleConnector.Config
				newConf := &google.Config{
					ClientID:                         oldConf.ClientID,
					ExistingClientIDSecret:           oldConf.ExistingClientIDSecret,
					ClientSecret:                     oldConf.ClientSecret,
					ExistingClientSecretSecret:       oldConf.ExistingClientSecretSecret,
					ServiceAccountJSON:               oldConf.ServiceAccountJSON,
					ExistingServiceAccountJSONSecret: oldConf.ExistingServiceAccountJSONSecret,
					ServiceAccountFilePath:           oldConf.ServiceAccountFilePath,
					AdminEmail:                       oldConf.AdminEmail,
				}
				if oldConf.HostedDomains != nil {
					domains := make([]string, 0)
					for _, v := range oldConf.HostedDomains {
						domains = append(domains, v)
					}
					oldConf.HostedDomains = domains
				}
				if oldConf.Groups != nil {
					groups := make([]string, 0)
					for _, v := range oldConf.Groups {
						groups = append(groups, v)
					}
					newConf.Groups = groups
				}
				conn.Config = newConf
			}
			newAuth.GoogleConnector = conn
		}
		if oldAuth.GithubConnector != nil {
			conn := &github.Connector{
				ID:   oldAuth.GithubConnector.ID,
				Name: oldAuth.GithubConnector.Name,
			}
			if oldAuth.GithubConnector.Config != nil {
				oldConf := oldAuth.GithubConnector.Config
				newConf := &github.Config{
					ClientID:                   oldConf.ClientID,
					ExistingClientIDSecret:     oldConf.ExistingClientIDSecret,
					ClientSecret:               oldConf.ClientSecret,
					ExistingClientSecretSecret: oldConf.ExistingClientSecretSecret,
					LoadAllGroups:              oldConf.LoadAllGroups,
					TeamNameField:              oldConf.TeamNameField,
					UseLoginAsID:               oldConf.UseLoginAsID,
				}
				if oldConf.Orgs != nil {
					orgs := make([]*github.Org, 0)
					for _, v := range oldConf.Orgs {
						org := &github.Org{
							Name:  v.Name,
							Teams: nil,
						}
						if v.Teams != nil {
							ts := make([]string, 0)
							for _, t := range v.Teams {
								ts = append(ts, t)
							}
						}
						orgs = append(orgs, org)
					}
					newConf.Orgs = orgs
				}
				conn.Config = newConf
			}
			newAuth.GithubConnector = conn
		}

		if oldAuth.GitlabConnector != nil {
			conn := &gitlab.Connector{
				ID:   oldAuth.GitlabConnector.ID,
				Name: oldAuth.GitlabConnector.Name,
			}
			if oldAuth.GitlabConnector.Config != nil {
				oldConf := oldAuth.GitlabConnector.Config
				newConf := &gitlab.Config{
					ClientID:                   oldConf.ClientID,
					ExistingClientIDSecret:     oldConf.ExistingClientIDSecret,
					ClientSecret:               oldConf.ClientSecret,
					ExistingClientSecretSecret: oldConf.ExistingClientSecretSecret,
					UseLoginAsID:               oldConf.UseLoginAsID,
					BaseURL:                    oldConf.BaseURL,
				}
				if oldConf.Groups != nil {
					groups := make([]string, 0)
					for _, v := range oldConf.Groups {
						groups = append(groups, v)
					}
					newConf.Groups = groups
				}
				conn.Config = newConf
			}
			newAuth.GitlabConnector = conn
		}
		if oldAuth.OIDC != nil {
			conn := &oidc.OIDC{
				Name:                       oldAuth.OIDC.Name,
				Issuer:                     oldAuth.OIDC.Issuer,
				ClientID:                   oldAuth.OIDC.ClientID,
				ExistingClientIDSecret:     oldAuth.OIDC.ExistingClientIDSecret,
				ClientSecret:               oldAuth.OIDC.ClientSecret,
				ExistingClientSecretSecret: oldAuth.OIDC.ExistingClientSecretSecret,
				RequestedIDTokenClaims:     nil,
			}
			if oldAuth.OIDC.RequestedScopes != nil {
				scopes := make([]string, 0)
				for _, v := range oldAuth.OIDC.RequestedScopes {
					scopes = append(scopes, v)
				}
				conn.RequestedScopes = scopes
			}
			if oldAuth.OIDC.RequestedIDTokenClaims != nil {
				claims := make(map[string]oidc.Claim, 0)
				for k, v := range oldAuth.OIDC.RequestedIDTokenClaims {
					claim := oidc.Claim{
						Essential: v.Essential,
					}
					if v.Values != nil {
						values := make([]string, 0)
						for _, value := range v.Values {
							values = append(values, value)
						}
						claim.Values = values
					}
					claims[k] = claim
				}
				conn.RequestedIDTokenClaims = claims
			}
			newAuth.OIDC = conn
		}
		new.Auth = &newAuth
	}

	return new
}
