//go:generate stringer -type=Protocol

package model

import "github.com/caos/infrop/internal/kinds/clusters/core/infra"

type HealthChecks struct {
	Port int64
	Path string
}

type Protocol int

const (
	Unknown Protocol = iota
	TCP
	UDP
)

/*type LoadBalancer struct {
	Pools []string
	//	Port         uint16
	//	Protocol     Protocol
	//	External     bool
	//	HealthChecks *HealthChecks
	HealthChecksPath string
}*/

type Pool struct {
	OSImage     string
	MinCPUCores int
	MinMemoryGB int
	StorageGB   int
}

type UserSpec struct {
	Verbose    bool
	RemoteUser string
	Project    string
	Region     string
	Zone       string
	Pools      map[string]*Pool
}

type Config struct {
}

type Current struct {
	infra.ProviderCurrent
}
