package orb

import "github.com/caos/orbiter/internal/operator/orbiter"

type Deps struct {
	Clusters  map[string]*orbiter.Tree
	Providers map[string]*orbiter.Tree
}

type DesiredV0 struct {
	Common *orbiter.Common `yaml:",inline"`
	Deps   Deps
}

type SecretsV0 struct {
	Common *orbiter.Common `yaml:",inline"`
	Deps   Deps
}

type Current struct {
	Common *orbiter.Common `yaml:",inline"`
	Deps   Deps
}
