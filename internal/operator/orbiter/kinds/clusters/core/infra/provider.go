package infra

import (
	"fmt"
	"io"
)

type Address struct {
	Location *string
	Port     uint16
}

func (a Address) String() string {
	return fmt.Sprintf("%s:%d", *a.Location, a.Port)
}

type ProviderCurrent interface {
	Pools() map[string]Pool
	Ingresses() map[string]*Address
}

type Ingress struct {
	Pools            []string
	HealthChecksPath string
}

type Pool interface {
	EnsureMembers() error
	GetMachines() (Machines, error)
	AddMachine() (Machine, error)
}

type Machine interface {
	ID() string
	IP() string
	Remove() error
	Execute(env map[string]string, stdin io.Reader, cmd string) ([]byte, error)
	WriteFile(path string, data io.Reader, permissions uint16) error
	ReadFile(path string, data io.Writer) error
}

type Machines []Machine

func (c Machines) ToChan() <-chan Machine {
	compChan := make(chan Machine)
	go func() {
		for _, comp := range c {
			compChan <- comp
		}
		close(compChan)
	}()
	return compChan
}

func (c Machines) String() string {
	list := ""
	for _, comp := range c {
		list += "|" + comp.ID()
	}
	if len(list) > 0 {
		list = list[1:]
	}
	return "(" + list + ")"
}

func (c Machines) Len() int           { return len(c) }
func (c Machines) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Machines) Less(i, j int) bool { return c[i].ID() < c[j].ID() }
