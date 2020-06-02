package grafana

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/admin"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/network"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/storage"
	"reflect"
)

type Grafana struct {
	Deploy             bool             `json:"deploy" yaml:"deploy"`
	Admin              *admin.Admin     `json:"admin,omitempty" yaml:"admin,omitempty"`
	Datasources        []*Datasource    `json:"datasources,omitempty" yaml:"datasources,omitempty"`
	DashboardProviders []*Provider      `json:"dashboardproviders,omitempty" yaml:"dashboardproviders,omitempty"`
	Storage            *storage.Spec    `json:"storage,omitempty" yaml:"storage,omitempty"`
	Network            *network.Network `json:"network,omitempty" yaml:"network,omitempty"`
	Auth               *auth.Auth       `json:"auth,omitempty" yaml:"auth,omitempty"`
	Plugins            []string         `json:"plugins,omitempty" yaml:"plugins,omitempty"`
}

func (x *Grafana) MarshalYAML() (interface{}, error) {
	type Alias Grafana
	return &Alias{
		Deploy:             x.Deploy,
		Admin:              admin.ClearEmpty(x.Admin),
		Datasources:        x.Datasources,
		DashboardProviders: x.DashboardProviders,
		Storage:            x.Storage,
		Network:            x.Network,
		Auth:               auth.ClearEmpty(x.Auth),
	}, nil
}

func ClearEmpty(x *Grafana) *Grafana {
	if x == nil {
		return nil
	}

	marshaled := Grafana{
		Deploy:             x.Deploy,
		Admin:              admin.ClearEmpty(x.Admin),
		Datasources:        x.Datasources,
		DashboardProviders: x.DashboardProviders,
		Storage:            x.Storage,
		Network:            x.Network,
		Auth:               auth.ClearEmpty(x.Auth),
	}
	if reflect.DeepEqual(marshaled, Grafana{}) {
		return &Grafana{}
	}
	return &marshaled
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
