package model

import (
	"github.com/caos/orbiter/internal/core/logging"
	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
)

type UserSpec struct {
	Destroyed bool
	Verbose   bool
	Versions  struct {
		Kubernetes string
		Orbiter     string
		Toolsop    string
	}
	Networking struct {
		DNSDomain   string
		Network     string
		ServiceCidr string
		PodCidr     string
	}
	ControlPlane *Pool
	Workers      map[string]*Pool
}

type Parameters struct {
	Logger           logging.Logger
	ID               string
	RepoURL          string
	RepoKey          string
	MasterKey        string
	OrbiterVersion    string
	CurrentFile      string
	SecretsFile      string
	SelfAbsolutePath []string
}

type Config struct {
	Spec      UserSpec
	Params    Parameters
	NodeAgent func(cmp infra.Compute) *operator.NodeAgentCurrent
}

type Pool struct {
	UpdatesDisabled bool
	Provider        string
	Nodes           int
	Pool            string
}

type Current struct {
	Status   string
	Computes map[string]*Compute `yaml:"computes"`
}

type Compute struct {
	Status    string
	Metadata  *ComputeMetadata
	Nodeagent *operator.NodeAgentCurrent `yaml:"-"`
}

type ComputeMetadata struct {
	Tier     Tier
	Provider string
	Pool     string
	Group    string `yaml:",omitempty"`
}

type Tier string

const (
	Controlplane Tier = "controlplane"
	Workers      Tier = "workers"
)
