package reconciling

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/k8s"
	"github.com/caos/orbos/internal/operator/boom/api/latest/network"
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling/auth"
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling/repository"
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
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run argocd on nodes
	Tolerations k8s.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	//Dex options
	Dex *CommonComponent `json:"dex,omitempty" yaml:"dex,omitempty"`
	//RepoServer options
	RepoServer *CommonComponent `json:"repoServer,omitempty" yaml:"repoServer,omitempty"`
	//Redis options
	Redis *CommonComponent `json:"redis,omitempty" yaml:"redis,omitempty"`
	//Controller options
	Controller *CommonComponent `json:"controller,omitempty" yaml:"controller,omitempty"`
	//Server options
	Server *CommonComponent `json:"server,omitempty" yaml:"server,omitempty"`
}

func (r *Reconciling) InitSecrets() {
	if r.Auth == nil {
		r.Auth = &auth.Auth{}
	}
	r.Auth.InitSecrets()
}

func (r *Reconciling) IsZero() bool {
	if !r.Deploy &&
		r.CustomImage == nil &&
		r.Network == nil &&
		(r.Auth == nil || r.Auth.IsZero()) &&
		r.Rbac == nil &&
		r.Repositories == nil &&
		r.Credentials == nil &&
		r.KnownHosts == nil &&
		r.NodeSelector == nil &&
		r.Tolerations == nil &&
		r.Dex == nil &&
		r.RepoServer == nil &&
		r.Redis == nil &&
		r.Controller == nil &&
		r.Server == nil {
		return true
	}

	return false
}

type CommonComponent struct {
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}

/*
	values.Dex.Tolerations = append(values.Dex.Tolerations, t)
	values.RepoServer.Tolerations = append(values.RepoServer.Tolerations, t)
	values.Redis.Tolerations = append(values.Redis.Tolerations, t)
	values.Controller.Tolerations = append(values.Controller.Tolerations, t)
	values.Server.Tolerations = append(values.Server.Tolerations, t)
*/

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
