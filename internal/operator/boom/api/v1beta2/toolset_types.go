package v1beta2

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/admin"
	monitoringauth "github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth"
	monitoringgeneric "github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth/Generic"
	monitoringgithub "github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth/Github"
	monitoringgitlab "github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth/Gitlab"
	monitoringgoogle "github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring/auth/Google"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling"
	reconcilingauth "github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/github"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/gitlab"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/google"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth/oidc"
	"github.com/caos/orbos/internal/secret"
)

type Metadata struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type ToolsetSpec struct {
	BoomVersion            string                   `json:"boomVersion,omitempty" yaml:"boomVersion,omitempty"`
	ForceApply             bool                     `json:"forceApply,omitempty" yaml:"forceApply,omitempty"`
	CurrentStateFolder     string                   `json:"currentStatePath,omitempty" yaml:"currentStatePath,omitempty"`
	PreApply               *PreApply                `json:"preApply,omitempty" yaml:"preApply,omitempty"`
	PostApply              *PostApply               `json:"postApply,omitempty" yaml:"postApply,omitempty"`
	MetricCollection       *MetricCollection        `json:"metricCollection,omitempty" yaml:"metricCollection,omitempty"`
	LogCollection          *LogCollection           `json:"logCollection,omitempty" yaml:"logCollection,omitempty"`
	NodeMetricsExporter    *NodeMetricsExporter     `json:"nodeMetricsExporter,omitempty" yaml:"nodeMetricsExporter,omitempty"`
	SystemdMetricsExporter *SystemdMetricsExporter  `json:"systemdMetricsExporter,omitempty" yaml:"systemdMetricsExporter,omitempty"`
	Monitoring             *monitoring.Monitoring   `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`
	APIGateway             *APIGateway              `json:"apiGateway,omitempty" yaml:"apiGateway,omitempty"`
	KubeMetricsExporter    *KubeMetricsExporter     `json:"kubeMetricsExporter,omitempty" yaml:"kubeMetricsExporter,omitempty"`
	Reconciling            *reconciling.Reconciling `json:"reconciling,omitempty" yaml:"reconciling,omitempty"`
	MetricsPersisting      *MetricsPersisting       `json:"metricsPersisting,omitempty" yaml:"metricsPersisting,omitempty"`
	LogsPersisting         *LogsPersisting          `json:"logsPersisting,omitempty" yaml:"logsPersisting,omitempty"`
}

type Toolset struct {
	APIVersion string       `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string       `json:"kind,omitempty" yaml:"kind,omitempty"`
	Metadata   *Metadata    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec       *ToolsetSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type ToolsetMetadata struct {
	APIVersion string    `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string    `json:"kind,omitempty" yaml:"kind,omitempty"`
	Metadata   *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

func (t *ToolsetSpec) MarshalYAML() (interface{}, error) {
	type Alias ToolsetSpec
	return &Alias{
		ForceApply:             t.ForceApply,
		CurrentStateFolder:     t.CurrentStateFolder,
		PreApply:               t.PreApply,
		PostApply:              t.PostApply,
		MetricCollection:       t.MetricCollection,
		LogCollection:          t.LogCollection,
		NodeMetricsExporter:    t.NodeMetricsExporter,
		SystemdMetricsExporter: t.SystemdMetricsExporter,
		Monitoring:             monitoring.ClearEmpty(t.Monitoring),
		APIGateway:             t.APIGateway,
		KubeMetricsExporter:    t.KubeMetricsExporter,
		Reconciling:            reconciling.ClearEmpty(t.Reconciling),
		MetricsPersisting:      t.MetricsPersisting,
		LogsPersisting:         t.LogsPersisting,
	}, nil
}

func (t *Toolset) InitSecretLists(masterkey string) error {
	if t.Spec.Reconciling != nil && t.Spec.Reconciling.Credentials != nil {
		for _, value := range t.Spec.Reconciling.Credentials {
			value.Username = secret.InitIfNil(value.Username, masterkey)
			if err := value.Username.Unmarshal(masterkey); err != nil {
				return err
			}
			value.Password = secret.InitIfNil(value.Password, masterkey)
			if err := value.Password.Unmarshal(masterkey); err != nil {
				return err
			}
			value.Certificate = secret.InitIfNil(value.Certificate, masterkey)
			if err := value.Certificate.Unmarshal(masterkey); err != nil {
				return err
			}
		}
	}

	if t.Spec.Reconciling != nil && t.Spec.Reconciling.Repositories != nil {
		for _, value := range t.Spec.Reconciling.Repositories {
			value.Username = secret.InitIfNil(value.Username, masterkey)
			if err := value.Username.Unmarshal(masterkey); err != nil {
				return err
			}
			value.Password = secret.InitIfNil(value.Password, masterkey)
			if err := value.Password.Unmarshal(masterkey); err != nil {
				return err
			}
			value.Certificate = secret.InitIfNil(value.Certificate, masterkey)
			if err := value.Certificate.Unmarshal(masterkey); err != nil {
				return err
			}
		}
	}

	if t.Spec.Reconciling != nil && t.Spec.Reconciling.CustomImage != nil && t.Spec.Reconciling.CustomImage.GopassStores != nil {
		for _, value := range t.Spec.Reconciling.CustomImage.GopassStores {
			value.SSHKey = secret.InitIfNil(value.SSHKey, masterkey)
			if err := value.SSHKey.Unmarshal(masterkey); err != nil {
				return err
			}
			value.GPGKey = secret.InitIfNil(value.GPGKey, masterkey)
			if err := value.GPGKey.Unmarshal(masterkey); err != nil {
				return err
			}
		}
	}
	return nil
}

func New(masterkey string) *Toolset {
	return &Toolset{
		Spec: &ToolsetSpec{
			Monitoring: &monitoring.Monitoring{
				Admin: &admin.Admin{
					Username: &secret.Secret{Masterkey: masterkey},
					Password: &secret.Secret{Masterkey: masterkey},
				},
				Auth: &monitoringauth.Auth{
					Google: &monitoringgoogle.Auth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					Github: &monitoringgithub.Auth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					Gitlab: &monitoringgitlab.Auth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					GenericOAuth: &monitoringgeneric.Auth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
				},
			},
			Reconciling: &reconciling.Reconciling{
				Auth: &reconcilingauth.Auth{
					OIDC: &oidc.OIDC{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					GithubConnector: &github.Connector{
						Config: &github.Config{
							ClientID:     &secret.Secret{Masterkey: masterkey},
							ClientSecret: &secret.Secret{Masterkey: masterkey},
						},
					},
					GitlabConnector: &gitlab.Connector{
						Config: &gitlab.Config{
							ClientID:     &secret.Secret{Masterkey: masterkey},
							ClientSecret: &secret.Secret{Masterkey: masterkey},
						},
					},
					GoogleConnector: &google.Connector{
						Config: &google.Config{
							ClientID:           &secret.Secret{Masterkey: masterkey},
							ClientSecret:       &secret.Secret{Masterkey: masterkey},
							ServiceAccountJSON: &secret.Secret{Masterkey: masterkey},
						},
					},
				},
			},
		},
	}
}

func ReplaceMasterkey(toolset *Toolset, masterkey string) *Toolset {
	old := toolset
	if old != nil && old.Spec != nil {
		if old.Spec.Monitoring != nil {
			if old.Spec.Monitoring.Admin != nil {
				if old.Spec.Monitoring.Admin.Username != nil {
					old.Spec.Monitoring.Admin.Username.Masterkey = masterkey
				}
				if old.Spec.Monitoring.Admin.Password != nil {
					old.Spec.Monitoring.Admin.Password.Masterkey = masterkey
				}
			}
			if old.Spec.Monitoring.Auth != nil {
				if old.Spec.Monitoring.Auth.Google != nil {
					if old.Spec.Monitoring.Auth.Google.ClientID != nil {
						old.Spec.Monitoring.Auth.Google.ClientID.Masterkey = masterkey
					}
					if old.Spec.Monitoring.Auth.Google.ClientSecret != nil {
						old.Spec.Monitoring.Auth.Google.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Monitoring.Auth.Github != nil {
					if old.Spec.Monitoring.Auth.Github.ClientID != nil {
						old.Spec.Monitoring.Auth.Github.ClientID.Masterkey = masterkey
					}
					if old.Spec.Monitoring.Auth.Github.ClientSecret != nil {
						old.Spec.Monitoring.Auth.Github.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Monitoring.Auth.Gitlab != nil {
					if old.Spec.Monitoring.Auth.Gitlab.ClientID != nil {
						old.Spec.Monitoring.Auth.Gitlab.ClientID.Masterkey = masterkey
					}
					if old.Spec.Monitoring.Auth.Gitlab.ClientSecret != nil {
						old.Spec.Monitoring.Auth.Gitlab.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Monitoring.Auth.GenericOAuth != nil {
					if old.Spec.Monitoring.Auth.GenericOAuth.ClientID != nil {
						old.Spec.Monitoring.Auth.GenericOAuth.ClientID.Masterkey = masterkey
					}
					if old.Spec.Monitoring.Auth.GenericOAuth.ClientSecret != nil {
						old.Spec.Monitoring.Auth.GenericOAuth.ClientSecret.Masterkey = masterkey
					}
				}
			}
		}
		if old.Spec.Reconciling != nil {
			if old.Spec.Reconciling.Auth != nil {
				if old.Spec.Reconciling.Auth.GoogleConnector != nil && old.Spec.Reconciling.Auth.GoogleConnector.Config != nil {
					if old.Spec.Reconciling.Auth.GoogleConnector.Config.ClientID != nil {
						old.Spec.Reconciling.Auth.GoogleConnector.Config.ClientID.Masterkey = masterkey
					}
					if old.Spec.Reconciling.Auth.GoogleConnector.Config.ClientSecret != nil {
						old.Spec.Reconciling.Auth.GoogleConnector.Config.ClientSecret.Masterkey = masterkey
					}
					if old.Spec.Reconciling.Auth.GoogleConnector.Config.ServiceAccountJSON != nil {
						old.Spec.Reconciling.Auth.GoogleConnector.Config.ServiceAccountJSON.Masterkey = masterkey
					}
				}
				if old.Spec.Reconciling.Auth.GithubConnector != nil && old.Spec.Reconciling.Auth.GithubConnector.Config != nil {
					if old.Spec.Reconciling.Auth.GithubConnector.Config.ClientID != nil {
						old.Spec.Reconciling.Auth.GithubConnector.Config.ClientID.Masterkey = masterkey
					}
					if old.Spec.Reconciling.Auth.GithubConnector.Config.ClientSecret != nil {
						old.Spec.Reconciling.Auth.GithubConnector.Config.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Reconciling.Auth.GitlabConnector != nil && old.Spec.Reconciling.Auth.GitlabConnector.Config != nil {
					if old.Spec.Reconciling.Auth.GitlabConnector.Config.ClientID != nil {
						old.Spec.Reconciling.Auth.GitlabConnector.Config.ClientID.Masterkey = masterkey
					}
					if old.Spec.Reconciling.Auth.GitlabConnector.Config.ClientSecret != nil {
						old.Spec.Reconciling.Auth.GitlabConnector.Config.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Reconciling.Auth.OIDC != nil {
					if old.Spec.Reconciling.Auth.OIDC.ClientID != nil {
						old.Spec.Reconciling.Auth.OIDC.ClientID.Masterkey = masterkey
					}
					if old.Spec.Reconciling.Auth.OIDC.ClientSecret != nil {
						old.Spec.Reconciling.Auth.OIDC.ClientSecret.Masterkey = masterkey
					}
				}
			}
		}
		if old.Spec.Reconciling != nil {
			if old.Spec.Reconciling.Credentials != nil {
				for _, value := range old.Spec.Reconciling.Credentials {
					if value.Username != nil {
						value.Username.Masterkey = masterkey
					}
					if value.Password != nil {
						value.Password.Masterkey = masterkey
					}
					if value.Certificate != nil {
						value.Certificate.Masterkey = masterkey
					}
				}
			}
			if old.Spec.Reconciling.Repositories != nil {
				for _, value := range old.Spec.Reconciling.Repositories {
					if value.Username != nil {
						value.Username.Masterkey = masterkey
					}
					if value.Password != nil {
						value.Password.Masterkey = masterkey
					}
					if value.Certificate != nil {
						value.Certificate.Masterkey = masterkey
					}
				}
			}
			if old.Spec.Reconciling.CustomImage != nil && old.Spec.Reconciling.CustomImage.GopassStores != nil {
				for _, value := range old.Spec.Reconciling.CustomImage.GopassStores {
					if value.SSHKey != nil {
						value.SSHKey.Masterkey = masterkey
					}
					if value.GPGKey != nil {
						value.GPGKey.Masterkey = masterkey
					}
				}
			}
		}
	}

	return old
}
