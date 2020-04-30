package kubernetes

import (
	"github.com/caos/orbos/internal/tree"
)

type CurrentCluster struct {
	Status   string
	Machines map[string]*Machine `yaml:"machines"`
}

type Current struct {
	Common  tree.Common `yaml:",inline"`
	Current *CurrentCluster
}

type Machine struct {
	Joined             bool
	Online             bool
	FirewallIsReady    bool
	NodeAgentIsRunning bool
	Metadata           MachineMetadata `yaml:",inline"`
}

type Versions struct {
	NodeAgent  string
	Kubernetes string
}

type MachineMetadata struct {
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
