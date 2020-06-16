package monitoring

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/admin"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/network"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/storage"
)

type Monitoring struct {
	Deploy             bool             `json:"deploy" yaml:"deploy"`
	Admin              *admin.Admin     `json:"admin,omitempty" yaml:"admin,omitempty"`
	Datasources        []*Datasource    `json:"datasources,omitempty" yaml:"datasources,omitempty"`
	DashboardProviders []*Provider      `json:"dashboardproviders,omitempty" yaml:"dashboardproviders,omitempty"`
	Storage            *storage.Spec    `json:"storage,omitempty" yaml:"storage,omitempty"`
	Network            *network.Network `json:"network,omitempty" yaml:"network,omitempty"`
	Auth               *auth.Auth       `json:"auth,omitempty" yaml:"auth,omitempty"`
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
