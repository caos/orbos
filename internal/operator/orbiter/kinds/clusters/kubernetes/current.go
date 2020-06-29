package kubernetes

import (
	"sync"

	"github.com/caos/orbos/internal/tree"
)

type CurrentCluster struct {
	Status   string
	Machines Machines
}

type Machines struct {
	// M is exported for yaml (de)serialization and not intended to be accessed by any other code outside this package
	M   map[string]*Machine `yaml:",inline"`
	mux sync.Mutex          `yaml:"-"`
}

func (m *Machines) Delete(id string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	delete(m.M, id)
}

func (m *Machines) Set(id string, machine *Machine) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.M == nil {
		m.M = make(map[string]*Machine)
	}

	m.M[id] = machine
}

type Current struct {
	Common  tree.Common `yaml:",inline"`
	Current *CurrentCluster
}

type Machine struct {
	Joined          bool
	Online          bool
	Ready           bool
	FirewallIsReady bool
	Metadata        MachineMetadata `yaml:",inline"`
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
