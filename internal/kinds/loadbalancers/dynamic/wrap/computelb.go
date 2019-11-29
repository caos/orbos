package wrap

import (
	"io"

	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
)

type cmpLB struct {
	original infra.Compute
	desire   func() error
}

func compute(compute infra.Compute, desire func() error) infra.Compute {
	return &cmpLB{
		original: compute,
		desire:   desire,
	}
}

func (c *cmpLB) ID() string {
	return c.original.ID()
}
func (c *cmpLB) InternalIP() (*string, error) {
	return c.original.InternalIP()
}
func (c *cmpLB) ExternalIP() (*string, error) {
	return c.original.ExternalIP()
}
func (c *cmpLB) Remove() error {
	err := c.original.Remove()
	if err != nil {
		return err
	}
	return c.desire()
}
func (c *cmpLB) Execute(env map[string]string, stdin io.Reader, cmd string) ([]byte, error) {
	return c.original.Execute(env, stdin, cmd)
}
func (c *cmpLB) WriteFile(path string, data io.Reader, permissions uint16) error {
	return c.original.WriteFile(path, data, permissions)
}
func (c *cmpLB) ReadFile(path string, data io.Writer) error {
	return c.original.ReadFile(path, data)
}
func (c *cmpLB) UseKeys(sec *operator.Secrets, privateKeyPaths ...string) error {
	return c.original.UseKeys(sec, privateKeyPaths...)
}
