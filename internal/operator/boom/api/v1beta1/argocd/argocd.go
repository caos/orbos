package argocd

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/auth"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd/repository"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/network"
	"github.com/caos/orbos/internal/secret"
	"reflect"
)

type Argocd struct {
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
	Enabled      bool           `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	GopassStores []*GopassStore `json:"gopassStores,omitempty" yaml:"gopassStores,omitempty"`
}

type GopassStore struct {
	SSHKey               *secret.Secret   `yaml:"sshKey,omitempty"`
	ExistingSSHKeySecret *secret.Existing `json:"existingSshKeySecret,omitempty" yaml:"existingSshKeySecret,omitempty"`
	GPGKey               *secret.Secret   `yaml:"gpgKey,omitempty"`
	ExistingGPGKeySecret *secret.Existing `json:"existingGpgKeySecret,omitempty" yaml:"existingGpgKeySecret,omitempty"`
	Directory            string           `json:"directory,omitempty" yaml:"directory,omitempty"`
	StoreName            string           `json:"storeName,omitempty" yaml:"storeName,omitempty"`
}

func ClearEmpty(x *Argocd) *Argocd {
	if x == nil {
		return nil
	}

	repos := make([]*repository.Repository, 0)
	for _, v := range x.Repositories {
		if p := repository.ClearEmpty(v); p != nil {
			repos = append(repos, p)
		}
	}

	creds := make([]*repository.Repository, 0)
	for _, v := range x.Credentials {
		if p := repository.ClearEmpty(v); p != nil {
			creds = append(creds, p)
		}
	}

	marshaled := Argocd{
		Deploy:       x.Deploy,
		CustomImage:  x.CustomImage,
		Network:      x.Network,
		Auth:         auth.ClearEmpty(x.Auth),
		Rbac:         x.Rbac,
		Repositories: repos,
		Credentials:  creds,
		KnownHosts:   x.KnownHosts,
	}

	if reflect.DeepEqual(marshaled, Argocd{}) {
		return &Argocd{}
	}
	return &marshaled
}
