package orb

import "github.com/caos/orbiter/internal/operator/orbiter"

type DesiredV0 struct {
	Common *orbiter.Common `yaml:",inline"`
	Spec   struct {
		Verbose bool
	}
	Clusters  map[string]*orbiter.Tree
	Providers map[string]*orbiter.Tree
}

type SecretsV0 struct {
	Common    *orbiter.Common `yaml:",inline"`
	Clusters  map[string]*orbiter.Tree
	Providers map[string]*orbiter.Tree
}

type Current struct {
	Common    *orbiter.Common `yaml:",inline"`
	Clusters  map[string]*orbiter.Tree
	Providers map[string]*orbiter.Tree
}
