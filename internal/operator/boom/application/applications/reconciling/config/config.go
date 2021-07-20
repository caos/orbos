package config

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling"
	"github.com/caos/orbos/internal/operator/boom/application/applications/reconciling/config/auth"
	"github.com/caos/orbos/internal/operator/boom/application/applications/reconciling/config/credential"
	"github.com/caos/orbos/internal/operator/boom/application/applications/reconciling/config/plugin"
	"github.com/caos/orbos/internal/operator/boom/application/applications/reconciling/config/repository"
	"github.com/caos/orbos/mntr"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Repositories            string `yaml:"repositories,omitempty"`
	Credentials             string `yaml:"repository.credentials,omitempty"`
	Connectors              string `yaml:"connectors,omitempty"`
	OIDC                    string `yaml:"oidc,omitempty"`
	ConfigManagementPlugins string `yaml:"configManagementPlugins,omitempty"`
}

func GetFromSpec(monitor mntr.Monitor, spec *reconciling.Reconciling) *Config {
	conf := &Config{}

	dexconfig := auth.GetDexConfigFromSpec(monitor, spec)
	data, err := yaml.Marshal(dexconfig)
	if err == nil {
		conf.Connectors = string(data)
	}
	repos := repository.GetFromSpec(monitor, spec)
	data2, err := yaml.Marshal(repos)
	if err == nil {
		conf.Repositories = string(data2)
	}

	creds := credential.GetFromSpec(monitor, spec)
	data3, err := yaml.Marshal(creds)
	if err == nil {
		conf.Credentials = string(data3)
	}

	oidc, err := auth.GetOIDC(spec.Auth)
	if err == nil && oidc != "" {
		conf.OIDC = oidc
	}

	if spec.CustomImage != nil {
		plugins := make([]*plugin.Plugin, 0)
		init := &plugin.Command{
			Command: []string{"gopass", "sync"},
		}
		generate := &plugin.Command{
			Command: []string{"sh", "-c"},
			Args:    []string{"kustomize build && ./secrets.yaml.sh "},
		}
		plugins = append(plugins, plugin.New("getSecrets", init, generate))

		pluginsYaml, err := yaml.Marshal(plugins)
		if err == nil {
			conf.ConfigManagementPlugins = string(pluginsYaml)
		}
	}

	return conf
}
