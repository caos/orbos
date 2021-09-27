package kubernetes

import (
	"sync"

	"github.com/caos/orbos/v5/pkg/tree"
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
	Updating        bool
	Rebooting       bool
	Ready           bool
	FirewallIsReady bool
	Unknown         bool
	Metadata        MachineMetadata `yaml:",inline"`
}

func (m *Machine) GetUpdating() bool {
	return m.Updating
}

func (m *Machine) SetUpdating(u bool) {
	m.Updating = u
}

func (m *Machine) GetJoined() bool {
	return m.Joined
}

func (m *Machine) SetJoined(j bool) {
	m.Joined = j
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
