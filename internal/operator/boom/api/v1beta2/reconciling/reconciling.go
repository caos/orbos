package reconciling

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/network"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/repository"
	"github.com/caos/orbos/internal/secret"
)

type Reconciling struct {
	Deploy       bool                     `json:"deploy" yaml:"deploy"`
	CustomImage  *CustomImage             `json:"customImage,omitempty" yaml:"customImage,omitempty"`
	Network      *network.Network         `json:"network,omitempty" yaml:"network,omitempty"`
	Auth         *auth.Auth               `json:"auth,omitempty" yaml:"auth,omitempty"`
	Rbac         *Rbac                    `json:"rbacConfig,omitempty" yaml:"rbacConfig,omitempty"`
	Repositories []*repository.Repository `json:"repositories,omitempty" yaml:"repositories,omitempty"`
	Credentials  []*repository.Repository `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	KnownHosts   []string                 `json:"knownHosts,omitempty" yaml:"knownHosts,omitempty"`
}

type Rbac struct {
	Csv     string   `json:"policy.csv,omitempty" yaml:"policy.csv,omitempty"`
	Default string   `json:"policy.default,omitempty" yaml:"policy.default,omitempty"`
	Scopes  []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

type CustomImage struct {
	Enabled         bool           `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	ImagePullSecret string         `json:"imagePullSecret,omitempty" yaml:"imagePullSecret,omitempty"`
	GopassStores    []*GopassStore `json:"gopassStores,omitempty" yaml:"gopassStores,omitempty"`
}

type GopassStore struct {
	SSHKey               *secret.Secret   `json:"sshKey,omitempty" yaml:"sshKey,omitempty"`
	ExistingSSHKeySecret *secret.Existing `json:"existingSshKeySecret,omitempty" yaml:"existingSshKeySecret,omitempty"`
	GPGKey               *secret.Secret   `json:"gpgKey,omitempty" yaml:"gpgKey,omitempty"`
	ExistingGPGKeySecret *secret.Existing `json:"existingGpgKeySecret,omitempty" yaml:"existingGpgKeySecret,omitempty"`
	Directory            string           `json:"directory,omitempty" yaml:"directory,omitempty"`
	StoreName            string           `json:"storeName,omitempty" yaml:"storeName,omitempty"`
}
