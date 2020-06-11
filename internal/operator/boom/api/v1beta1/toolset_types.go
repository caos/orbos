package v1beta1

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd"
	argocdauth "github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth/github"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth/gitlab"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth/google"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth/oidc"
	grafana "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/admin"
	grafanaauth "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth"
	grafanageneric "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Generic"
	grafanagithub "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Github"
	grafanagitlab "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Gitlab"
	grafanagoogle "github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana/auth/Google"
	"github.com/caos/orbos/internal/secret"
)

type Metadata struct {
	//Name of the overlay
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Namespace as information, has currently no influence for the applied resources
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// ToolsetSpec: BOOM reconciles itself if a boomVersion is defined, if no boomVersion is defined there is no reconciling.
type ToolsetSpec struct {
	//Version of BOOM which should be reconciled
	BoomVersion string `json:"boomVersion,omitempty" yaml:"boomVersion,omitempty"`
	//Relative folder path where the currentstate is written to
	ForceApply bool `json:"forceApply,omitempty" yaml:"forceApply,omitempty"`
	//Flag if --force should be used by apply of resources
	CurrentStateFolder string `json:"currentStatePath,omitempty" yaml:"currentStatePath,omitempty"`
	//Spec for the yaml-files applied before the applications, for example used secrets
	PreApply *Apply `json:"preApply,omitempty" yaml:"preApply,omitempty"`
	//Spec for the yaml-files applied after the applications, for example additional crds for the applications
	PostApply *Apply `json:"postApply,omitempty" yaml:"postApply,omitempty"`
	//Spec for the Prometheus-Operator
	PrometheusOperator *PrometheusOperator `json:"prometheus-operator,omitempty" yaml:"prometheus-operator,omitempty"`
	//Spec for the Banzaicloud Logging-Operator
	LoggingOperator *LoggingOperator `json:"logging-operator,omitempty" yaml:"logging-operator,omitempty"`
	//Spec for the Prometheus-Node-Exporter
	PrometheusNodeExporter *PrometheusNodeExporter `json:"prometheus-node-exporter,omitempty" yaml:"prometheus-node-exporter,omitempty"`
	//Spec for the Prometheus-Systemd-Exporter
	PrometheusSystemdExporter *PrometheusSystemdExporter `json:"prometheus-systemd-exporter,omitempty" yaml:"prometheus-systemd-exporter,omitempty"`
	//Spec for the Grafana
	Grafana *grafana.Grafana `json:"grafana,omitempty" yaml:"grafana,omitempty"`
	//Spec for the Ambassador
	Ambassador *Ambassador `json:"ambassador,omitempty" yaml:"ambassador,omitempty"`
	//Spec for the Kube-State-Metrics
	KubeStateMetrics *KubeStateMetrics `json:"kube-state-metrics,omitempty" yaml:"kube-state-metrics,omitempty"`
	//Spec for the Argo-CD
	Argocd *argocd.Argocd `json:"argocd,omitempty" yaml:"argocd,omitempty"`
	//Spec for the Prometheus instance
	Prometheus *Prometheus `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
	//Spec for the Loki instance
	Loki *Loki `json:"loki,omitempty" yaml:"loki,omitempty"`
}

type Toolset struct {
	//Version of the used API
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	//Kind for the standard CRD
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
	//Metadata for the CRD
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	//Specification for the Toolset
	Spec *ToolsetSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type ToolsetMetadata struct {
	APIVersion string    `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string    `json:"kind,omitempty" yaml:"kind,omitempty"`
	Metadata   *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

func (t *ToolsetSpec) MarshalYAML() (interface{}, error) {
	type Alias ToolsetSpec
	return &Alias{
		ForceApply:                t.ForceApply,
		CurrentStateFolder:        t.CurrentStateFolder,
		PreApply:                  t.PreApply,
		PostApply:                 t.PostApply,
		PrometheusOperator:        t.PrometheusOperator,
		LoggingOperator:           t.LoggingOperator,
		PrometheusNodeExporter:    t.PrometheusNodeExporter,
		PrometheusSystemdExporter: t.PrometheusSystemdExporter,
		Grafana:                   grafana.ClearEmpty(t.Grafana),
		Ambassador:                t.Ambassador,
		KubeStateMetrics:          t.KubeStateMetrics,
		Argocd:                    argocd.ClearEmpty(t.Argocd),
		Prometheus:                t.Prometheus,
		Loki:                      t.Loki,
	}, nil
}

func (t *Toolset) InitSecretLists(masterkey string) error {
	if t.Spec.Argocd != nil && t.Spec.Argocd.Credentials != nil {
		for _, value := range t.Spec.Argocd.Credentials {
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

	if t.Spec.Argocd != nil && t.Spec.Argocd.Repositories != nil {
		for _, value := range t.Spec.Argocd.Repositories {
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

	if t.Spec.Argocd != nil && t.Spec.Argocd.CustomImage != nil && t.Spec.Argocd.CustomImage.GopassStores != nil {
		for _, value := range t.Spec.Argocd.CustomImage.GopassStores {
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
			Grafana: &grafana.Grafana{
				Admin: &admin.Admin{
					Username: &secret.Secret{Masterkey: masterkey},
					Password: &secret.Secret{Masterkey: masterkey},
				},
				Auth: &grafanaauth.Auth{
					Google: &grafanagoogle.Auth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					Github: &grafanagithub.Auth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					Gitlab: &grafanagitlab.Auth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
					GenericOAuth: &grafanageneric.Auth{
						ClientID:     &secret.Secret{Masterkey: masterkey},
						ClientSecret: &secret.Secret{Masterkey: masterkey},
					},
				},
			},
			Argocd: &argocd.Argocd{
				Auth: &argocdauth.Auth{
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
		if old.Spec.Grafana != nil {
			if old.Spec.Grafana.Admin != nil {
				if old.Spec.Grafana.Admin.Username != nil {
					old.Spec.Grafana.Admin.Username.Masterkey = masterkey
				}
				if old.Spec.Grafana.Admin.Password != nil {
					old.Spec.Grafana.Admin.Password.Masterkey = masterkey
				}
			}
			if old.Spec.Grafana.Auth != nil {
				if old.Spec.Grafana.Auth.Google != nil {
					if old.Spec.Grafana.Auth.Google.ClientID != nil {
						old.Spec.Grafana.Auth.Google.ClientID.Masterkey = masterkey
					}
					if old.Spec.Grafana.Auth.Google.ClientSecret != nil {
						old.Spec.Grafana.Auth.Google.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Grafana.Auth.Github != nil {
					if old.Spec.Grafana.Auth.Github.ClientID != nil {
						old.Spec.Grafana.Auth.Github.ClientID.Masterkey = masterkey
					}
					if old.Spec.Grafana.Auth.Github.ClientSecret != nil {
						old.Spec.Grafana.Auth.Github.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Grafana.Auth.Gitlab != nil {
					if old.Spec.Grafana.Auth.Gitlab.ClientID != nil {
						old.Spec.Grafana.Auth.Gitlab.ClientID.Masterkey = masterkey
					}
					if old.Spec.Grafana.Auth.Gitlab.ClientSecret != nil {
						old.Spec.Grafana.Auth.Gitlab.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Grafana.Auth.GenericOAuth != nil {
					if old.Spec.Grafana.Auth.GenericOAuth.ClientID != nil {
						old.Spec.Grafana.Auth.GenericOAuth.ClientID.Masterkey = masterkey
					}
					if old.Spec.Grafana.Auth.GenericOAuth.ClientSecret != nil {
						old.Spec.Grafana.Auth.GenericOAuth.ClientSecret.Masterkey = masterkey
					}
				}
			}
		}
		if old.Spec.Argocd != nil {
			if old.Spec.Argocd.Auth != nil {
				if old.Spec.Argocd.Auth.GoogleConnector != nil && old.Spec.Argocd.Auth.GoogleConnector.Config != nil {
					if old.Spec.Argocd.Auth.GoogleConnector.Config.ClientID != nil {
						old.Spec.Argocd.Auth.GoogleConnector.Config.ClientID.Masterkey = masterkey
					}
					if old.Spec.Argocd.Auth.GoogleConnector.Config.ClientSecret != nil {
						old.Spec.Argocd.Auth.GoogleConnector.Config.ClientSecret.Masterkey = masterkey
					}
					if old.Spec.Argocd.Auth.GoogleConnector.Config.ServiceAccountJSON != nil {
						old.Spec.Argocd.Auth.GoogleConnector.Config.ServiceAccountJSON.Masterkey = masterkey
					}
				}
				if old.Spec.Argocd.Auth.GithubConnector != nil && old.Spec.Argocd.Auth.GithubConnector.Config != nil {
					if old.Spec.Argocd.Auth.GithubConnector.Config.ClientID != nil {
						old.Spec.Argocd.Auth.GithubConnector.Config.ClientID.Masterkey = masterkey
					}
					if old.Spec.Argocd.Auth.GithubConnector.Config.ClientSecret != nil {
						old.Spec.Argocd.Auth.GithubConnector.Config.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Argocd.Auth.GitlabConnector != nil && old.Spec.Argocd.Auth.GitlabConnector.Config != nil {
					if old.Spec.Argocd.Auth.GitlabConnector.Config.ClientID != nil {
						old.Spec.Argocd.Auth.GitlabConnector.Config.ClientID.Masterkey = masterkey
					}
					if old.Spec.Argocd.Auth.GitlabConnector.Config.ClientSecret != nil {
						old.Spec.Argocd.Auth.GitlabConnector.Config.ClientSecret.Masterkey = masterkey
					}
				}
				if old.Spec.Argocd.Auth.OIDC != nil {
					if old.Spec.Argocd.Auth.OIDC.ClientID != nil {
						old.Spec.Argocd.Auth.OIDC.ClientID.Masterkey = masterkey
					}
					if old.Spec.Argocd.Auth.OIDC.ClientSecret != nil {
						old.Spec.Argocd.Auth.OIDC.ClientSecret.Masterkey = masterkey
					}
				}
			}
		}
		if old.Spec.Argocd != nil {
			if old.Spec.Argocd.Credentials != nil {
				for _, value := range old.Spec.Argocd.Credentials {
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
			if old.Spec.Argocd.Repositories != nil {
				for _, value := range old.Spec.Argocd.Repositories {
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
			if old.Spec.Argocd.CustomImage != nil && old.Spec.Argocd.CustomImage.GopassStores != nil {
				for _, value := range old.Spec.Argocd.CustomImage.GopassStores {
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
