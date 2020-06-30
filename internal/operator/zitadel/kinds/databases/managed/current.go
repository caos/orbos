package managed

import (
	"github.com/caos/orbos/internal/tree"
)

type Current struct {
	Common  *tree.Common `yaml:",inline"`
	Current struct {
		URL   string
		Port  string
		Users []string
	}
}

func (c *Current) GetURL() string {
	return c.Current.URL
}

func (c *Current) GetPort() string {
	return c.Current.URL
}

func (c *Current) GetUsers() []string {
	return c.Current.Users
}
