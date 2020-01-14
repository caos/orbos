package kubernetes

import (
	"github.com/caos/orbiter/internal/core/operator/orbiter"
)

type CurrentCluster struct {
	Status   string
	Computes map[string]*Compute `yaml:"computes"`
}

type Current struct {
	Common  orbiter.Common `yaml:",inline"`
	Current CurrentCluster
	Deps    map[string]*orbiter.Tree
}

type Compute struct {
	Status    string
	Metadata  ComputeMetadata
	Nodeagent orbiter.NodeAgentCurrent `yaml:"-"`
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
