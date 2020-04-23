package v1beta1

import "github.com/caos/orbiter/internal/secret"

type Argocd struct {
	Deploy       bool                `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	CustomImage  *ArgocdCustomImage  `json:"customImage,omitempty" yaml:"customImage,omitempty"`
	Network      *Network            `json:"network,omitempty" yaml:"network,omitempty"`
	Auth         *ArgocdAuth         `json:"auth,omitempty" yaml:"auth,omitempty"`
	Rbac         *Rbac               `json:"rbacConfig,omitempty" yaml:"rbacConfig,omitempty"`
	Repositories []*ArgocdRepository `json:"repositories,omitempty" yaml:"repositories,omitempty"`
	KnownHosts   []string            `json:"knownHosts,omitempty" yaml:"knownHosts,omitempty"`
}

type Rbac struct {
	Csv     string   `json:"policy.csv,omitempty" yaml:"policy.csv,omitempty"`
	Default string   `json:"policy.default,omitempty" yaml:"policy.default,omitempty"`
	Scopes  []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

type ArgocdRepository struct {
	Name                      string           `json:"name,omitempty" yaml:"name,omitempty"`
	URL                       string           `json:"url,omitempty" yaml:"url,omitempty"`
	Username                  *secret.Secret   `yaml:"username,omitempty"`
	ExistingUsernameSecret    *secret.Existing `json:"existingUsernameSecret,omitempty" yaml:"existingUsernameSecret,omitempty"`
	Password                  *secret.Secret   `yaml:"password,omitempty"`
	ExistingPasswordSecret    *secret.Existing `json:"existingPasswordSecret,omitempty" yaml:"existingPasswordSecret,omitempty"`
	Certificate               *secret.Secret   `yaml:"certificate,omitempty"`
	ExistingCertificateSecret *secret.Existing `json:"existingCertificateSecret,omitempty" yaml:"existingCertificateSecret,omitempty"`
}

type ArgocdCustomImage struct {
	Enabled         bool                 `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	ImagePullSecret string               `json:"imagePullSecret,omitempty" yaml:"imagePullSecret,omitempty"`
	GopassStores    []*ArgocdGopassStore `json:"gopassStores,omitempty" yaml:"gopassStores,omitempty"`
}

type ArgocdGopassStore struct {
	SSHKey               *secret.Secret   `yaml:"sshKey,omitempty"`
	ExistingSSHKeySecret *secret.Existing `json:"existingSshKeySecret,omitempty" yaml:"existingSshKeySecret,omitempty"`
	GPGKey               *secret.Secret   `yaml:"gpgKey,omitempty"`
	ExistingGPGKeySecret *secret.Existing `json:"existingGpgKeySecret,omitempty" yaml:"existingGpgKeySecret,omitempty"`
	Directory            string           `json:"directory,omitempty" yaml:"directory,omitempty"`
	StoreName            string           `json:"storeName,omitempty" yaml:"storeName,omitempty"`
}

type ArgocdAuth struct {
	OIDC            *ArgocdOIDC            `json:"oidc,omitempty" yaml:"oidc,omitempty"`
	GithubConnector *ArgocdGithubConnector `json:"github,omitempty" yaml:"github,omitempty"`
	GitlabConnector *ArgocdGitlabConnector `json:"gitlab,omitempty" yaml:"gitlab,omitempty"`
	GoogleConnector *ArgocdGoogleConnector `json:"google,omitempty" yaml:"google,omitempty"`
}

type ArgocdOIDC struct {
	Name                       string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Issuer                     string                 `json:"issuer,omitempty" yaml:"issuer,omitempty"`
	ClientID                   *secret.Secret         `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing       `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret         `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing       `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	RequestedScopes            []string               `json:"requestedScopes,omitempty" yaml:"requestedScopes,omitempty"`
	RequestedIDTokenClaims     map[string]ArgocdClaim `json:"requestedIDTokenClaims,omitempty" yaml:"requestedIDTokenClaims,omitempty"`
}

type ArgocdClaim struct {
	Essential bool     `json:"essential,omitempty" yaml:"essential,omitempty"`
	Values    []string `json:"values,omitempty" yaml:"values,omitempty"`
}

type ArgocdGithubConnector struct {
	ID     string              `json:"id,omitempty" yaml:"id,omitempty"`
	Name   string              `json:"name,omitempty" yaml:"name,omitempty"`
	Config *ArgocdGithubConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

type ArgocdGithubConfig struct {
	ClientID                   *secret.Secret     `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing   `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret     `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing   `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	Orgs                       []*ArgocdGithubOrg `json:"orgs,omitempty" yaml:"orgs,omitempty"`
	LoadAllGroups              bool               `json:"loadAllGroups,omitempty" yaml:"loadAllGroups,omitempty"`
	TeamNameField              string             `json:"teamNameField,omitempty" yaml:"teamNameField,omitempty"`
	UseLoginAsID               bool               `json:"useLoginAsID,omitempty" yaml:"useLoginAsID,omitempty"`
}

type ArgocdGithubOrg struct {
	Name  string   `json:"name,omitempty" yaml:"name,omitempty"`
	Teams []string `json:"teams,omitempty" yaml:"teams,omitempty"`
}

type ArgocdGitlabConnector struct {
	ID     string              `json:"id,omitempty" yaml:"id,omitempty"`
	Name   string              `json:"name,omitempty" yaml:"name,omitempty"`
	Config *ArgocdGitlabConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

type ArgocdGitlabConfig struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	BaseURL                    string           `json:"baseURL,omitempty" yaml:"baseURL,omitempty"`
	Groups                     []string         `json:"groups,omitempty" yaml:"groups,omitempty"`
	UseLoginAsID               bool             `json:"useLoginAsID,omitempty" yaml:"useLoginAsID,omitempty"`
}

type ArgocdGoogleConnector struct {
	ID     string              `json:"id,omitempty" yaml:"id,omitempty"`
	Name   string              `json:"name,omitempty" yaml:"name,omitempty"`
	Config *ArgocdGoogleConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

type ArgocdGoogleConfig struct {
	ClientID                         *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret           *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret                     *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret       *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	HostedDomains                    []string         `json:"hostedDomains,omitempty" yaml:"hostedDomains,omitempty"`
	Groups                           []string         `json:"groups,omitempty" yaml:"groups,omitempty"`
	ServiceAccountJSON               *secret.Secret   `yaml:"serviceAccountJSON,omitempty"`
	ExistingServiceAccountJSONSecret *secret.Existing `json:"existingServiceAccountJSONSecret,omitempty" yaml:"existingServiceAccountJSONSecret,omitempty"`
	ServiceAccountFilePath           string           `json:"serviceAccountFilePath,omitempty" yaml:"serviceAccountFilePath,omitempty"`
	AdminEmail                       string           `json:"adminEmail,omitempty" yaml:"adminEmail,omitempty"`
}
