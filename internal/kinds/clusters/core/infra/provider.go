package infra

import (
	"fmt"
	"io"
)

type Address struct {
	Location string
	Port     uint16
}

func (a Address) String() string {
	return fmt.Sprintf("%s:%d", a.Location, a.Port)
}

type ProviderCurrent interface {
	Pools() map[string]Pool
	Ingresses() map[string]Address
	Cleanupped() <-chan error
}

type Ingress struct {
	Pools            []string
	HealthChecksPath string
}

type Pool interface {
	EnsureMembers() error
	GetComputes(active bool) (Computes, error)
	AddCompute() (Compute, error)
}

type Compute interface {
	ID() string
	IP() string
	Remove() error
	Execute(env map[string]string, stdin io.Reader, cmd string) ([]byte, error)
	WriteFile(path string, data io.Reader, permissions uint16) error
	ReadFile(path string, data io.Writer) error
	UseKey(keys ...[]byte) error
}

type Computes []Compute

func (c Computes) ToChan() <-chan Compute {
	compChan := make(chan Compute)
	go func() {
		for _, comp := range c {
			compChan <- comp
		}
		close(compChan)
	}()
	return compChan
}

func (c Computes) String() string {
	list := ""
	for _, comp := range c {
		list += "|" + comp.ID()
	}
	if len(list) > 0 {
		list = list[1:]
	}
	return "(" + list + ")"
}

func (c Computes) Len() int           { return len(c) }
func (c Computes) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Computes) Less(i, j int) bool { return c[i].ID() < c[j].ID() }
