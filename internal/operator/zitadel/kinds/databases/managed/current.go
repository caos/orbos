package managed

import (
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/tree"
)

type Current struct {
	Common  *tree.Common `yaml:",inline"`
	Current struct {
		URL       string
		Port      string
		ReadyFunc zitadel.EnsureFunc
	}
}

func (c *Current) GetURL() string {
	return c.Current.URL
}

func (c *Current) GetPort() string {
	return c.Current.Port
}

func (c *Current) GetReadyQuery() zitadel.EnsureFunc {
	return c.Current.ReadyFunc
}
