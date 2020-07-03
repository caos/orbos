package reconciling

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/network"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/auth"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling/repository"
	"github.com/caos/orbos/internal/secret"
)

type Reconciling struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Use of custom argocd-image which includes gopass
	//@default: false
	CustomImage *CustomImage `json:"customImage,omitempty" yaml:"customImage,omitempty"`
	//Network configuration, used for SSO and external access
	Network *network.Network `json:"network,omitempty" yaml:"network,omitempty"`
	//Authorization and Authentication configuration for SSO
	Auth *auth.Auth `json:"auth,omitempty" yaml:"auth,omitempty"`
	//Configuration for RBAC in argocd
	Rbac *Rbac `json:"rbacConfig,omitempty" yaml:"rbacConfig,omitempty"`
	//Repositories used by argocd
	Repositories []*repository.Repository `json:"repositories,omitempty" yaml:"repositories,omitempty"`
	//Credentials used by argocd
	Credentials []*repository.Repository `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	//List of known_hosts as strings for argocd
	KnownHosts []string `json:"knownHosts,omitempty" yaml:"knownHosts,omitempty"`
}

type Rbac struct {
	//Attribute policy.csv which goes into configmap argocd-rbac-cm
	Csv string `json:"policy.csv,omitempty" yaml:"policy.csv,omitempty"`
	//Attribute policy.default which goes into configmap argocd-rbac-cm
	Default string `json:"policy.default,omitempty" yaml:"policy.default,omitempty"`
	//List of scopes which go into configmap argocd-rbac-cm
	Scopes []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

type CustomImage struct {
	//Flag if custom argocd-image should get used with gopass
	Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	//Name of used imagePullSecret to pull customImage
	ImagePullSecret string `json:"imagePullSecret,omitempty" yaml:"imagePullSecret,omitempty"`
	//List of gopass stores which should get cloned by argocd on startup
	GopassStores []*GopassStore `json:"gopassStores,omitempty" yaml:"gopassStores,omitempty"`
}

type GopassStore struct {
	SSHKey *secret.Secret `yaml:"sshKey,omitempty"`
	//Existing secret with ssh-key to clone the repository for gopass
	ExistingSSHKeySecret *secret.Existing `json:"existingSshKeySecret,omitempty" yaml:"existingSshKeySecret,omitempty"`
	GPGKey               *secret.Secret   `yaml:"gpgKey,omitempty"`
	//Existing secret with gpg-key to decode the repository for gopass
	ExistingGPGKeySecret *secret.Existing `json:"existingGpgKeySecret,omitempty" yaml:"existingGpgKeySecret,omitempty"`
	//URL to repository for gopass store
	Directory string `json:"directory,omitempty" yaml:"directory,omitempty"`
	//Name of the gopass store
	StoreName string `json:"storeName,omitempty" yaml:"storeName,omitempty"`
}
