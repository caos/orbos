package v1beta1

import (
	"github.com/caos/orbiter/internal/operator/boom/api/v1beta1/storage"
	"github.com/caos/orbiter/internal/secret"
)

type Grafana struct {
	Deploy             bool          `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	Admin              *Admin        `json:"admin,omitempty" yaml:"admin,omitempty"`
	Datasources        []*Datasource `json:"datasources,omitempty" yaml:"datasources,omitempty"`
	DashboardProviders []*Provider   `json:"dashboardproviders,omitempty" yaml:"dashboardproviders,omitempty"`
	Storage            *storage.Spec `json:"storage,omitempty" yaml:"storage,omitempty"`
	Network            *Network      `json:"network,omitempty" yaml:"network,omitempty"`
	Auth               *GrafanaAuth  `json:"auth,omitempty" yaml:"auth,omitempty"`
}

type Admin struct {
	Username       *secret.Secret           `yaml:"username,omitempty"`
	Password       *secret.Secret           `yaml:"password,omitempty"`
	ExistingSecret *secret.ExistingIDSecret `json:"existingSecret,omitempty" yaml:"existingSecret,omitempty"`
}

type Datasource struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Type      string `json:"type,omitempty" yaml:"type,omitempty"`
	Url       string `json:"url,omitempty" yaml:"url,omitempty"`
	Access    string `json:"access,omitempty" yaml:"access,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty" yaml:"isDefault,omitempty"`
}

type Provider struct {
	ConfigMaps []string `json:"configMaps,omitempty" yaml:"configMaps,omitempty"`
	Folder     string   `json:"folder,omitempty" yaml:"folder,omitempty"`
}

type GrafanaAuth struct {
	Google       *GrafanaGoogleAuth   `json:"google,omitempty" yaml:"google,omitempty"`
	Github       *GrafanaGithubAuth   `json:"github,omitempty" yaml:"github,omitempty"`
	Gitlab       *GrafanaGitlabAuth   `json:"gitlab,omitempty" yaml:"gitlab,omitempty"`
	GenericOAuth *GrafanaGenericOAuth `json:"genericOAuth,omitempty" yaml:"genericOAuth,omitempty"`
}

type GrafanaGoogleAuth struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	AllowedDomains             []string         `json:"allowedDomains,omitempty" yaml:"allowedDomains,omitempty"`
}

type GrafanaGithubAuth struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	AllowedOrganizations       []string         `json:"allowedOrganizations,omitempty" yaml:"allowedOrganizations,omitempty"`
	TeamIDs                    []string         `json:"teamIDs,omitempty" yaml:"teamIDs,omitempty"`
}

type GrafanaGitlabAuth struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	AllowedGroups              []string         `json:"allowedGroups,omitempty" yaml:"allowedGroups,omitempty"`
}

type GrafanaGenericOAuth struct {
	ClientID                   *secret.Secret   `yaml:"clientID,omitempty"`
	ExistingClientIDSecret     *secret.Existing `json:"existingClientIDSecret,omitempty" yaml:"existingClientIDSecret,omitempty"`
	ClientSecret               *secret.Secret   `yaml:"clientSecret,omitempty"`
	ExistingClientSecretSecret *secret.Existing `json:"existingClientSecretSecret,omitempty" yaml:"existingClientSecretSecret,omitempty"`
	Scopes                     []string         `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	AuthURL                    string           `json:"authURL,omitempty" yaml:"authURL,omitempty"`
	TokenURL                   string           `json:"tokenURL,omitempty" yaml:"tokenURL,omitempty"`
	APIURL                     string           `json:"apiURL,omitempty" yaml:"apiURL,omitempty"`
	AllowedDomains             []string         `json:"allowedDomains,omitempty" yaml:"allowedDomains,omitempty"`
}
