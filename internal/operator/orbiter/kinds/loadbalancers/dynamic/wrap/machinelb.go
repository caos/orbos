package wrap

import (
	"io"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
)

type cmpLB struct {
	original infra.Machine
	desire   func() error
}

func machine(machine infra.Machine, desire func() error) infra.Machine {
	return &cmpLB{
		original: machine,
		desire:   desire,
	}
}

func (c *cmpLB) ID() string {
	return c.original.ID()
}
func (c *cmpLB) IP() string {
	return c.original.IP()
}

func (c *cmpLB) Remove() error {
	err := c.original.Remove()
	if err != nil {
		return err
	}
	return c.desire()
}
func (c *cmpLB) Execute(stdin io.Reader, cmd string) ([]byte, error) {
	return c.original.Execute(stdin, cmd)
}
func (c *cmpLB) Shell(env map[string]string) error {
	return c.original.Shell(env)
}
func (c *cmpLB) WriteFile(path string, data io.Reader, permissions uint16) error {
	return c.original.WriteFile(path, data, permissions)
}
func (c *cmpLB) ReadFile(path string, data io.Writer) error {
	return c.original.ReadFile(path, data)
}
