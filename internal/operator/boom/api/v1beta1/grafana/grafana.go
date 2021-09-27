package grafana

import (
	"github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/grafana/admin"
	"github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/grafana/auth"
	"github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/network"
	"github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/storage"
)

type Grafana struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Spec for the definition of the admin account
	Admin *admin.Admin `json:"admin,omitempty" yaml:"admin,omitempty"`
	//Spec for additional datasources
	Datasources []*Datasource `json:"datasources,omitempty" yaml:"datasources,omitempty"`
	//Spec for additional Dashboardproviders
	DashboardProviders []*Provider `json:"dashboardproviders,omitempty" yaml:"dashboardproviders,omitempty"`
	//Spec to define how the persistence should be handled
	Storage *storage.Spec `json:"storage,omitempty" yaml:"storage,omitempty"`
	//Network configuration, used for SSO and external access
	Network *network.Network `json:"network,omitempty" yaml:"network,omitempty"`
	//Authorization and Authentication configuration for SSO
	Auth *auth.Auth `json:"auth,omitempty" yaml:"auth,omitempty"`
	//List of plugins which get added to the grafana instance
	Plugins []string `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
}

func (g *Grafana) InitSecrets() {
	if g.Admin == nil {
		g.Admin = &admin.Admin{}
	}
	g.Admin.InitSecrets()
	if g.Auth == nil {
		g.Auth = &auth.Auth{}
	}
	g.Auth.InitSecrets()
}

func (m *Grafana) IsZero() bool {
	if !m.Deploy &&
		(m.Admin == nil || m.Admin.IsZero()) &&
		(m.Auth == nil || m.Auth.IsZero()) &&
		m.Datasources == nil &&
		m.DashboardProviders == nil &&
		m.Storage == nil &&
		m.Network == nil &&
		m.Plugins == nil &&
		m.NodeSelector == nil {
		return true
	}

	return false
}

type Datasource struct {
	//Name of the datasource
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Type of the datasource (for example prometheus)
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	//URL to the datasource
	Url string `json:"url,omitempty" yaml:"url,omitempty"`
	//Access defintion of the datasource
	Access string `json:"access,omitempty" yaml:"access,omitempty"`
	//Boolean if datasource should be used as default
	IsDefault bool `json:"isDefault,omitempty" yaml:"isDefault,omitempty"`
}

type Provider struct {
	//ConfigMaps in which the dashboards are stored
	ConfigMaps []string `json:"configMaps,omitempty" yaml:"configMaps,omitempty"`
	//Local folder in which the dashboards are mounted
	Folder string `json:"folder,omitempty" yaml:"folder,omitempty"`
}
