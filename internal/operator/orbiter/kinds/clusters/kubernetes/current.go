package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/orbiter"
)

type CurrentCluster struct {
	Status   string
	Machines map[string]*Machine `yaml:"machines"`
}

type Current struct {
	Common  orbiter.Common `yaml:",inline"`
	Current CurrentCluster
}

type Machine struct {
	Status   string
	Metadata MachineMetadata `yaml:",inline"`
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
