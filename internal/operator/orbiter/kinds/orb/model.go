package orb

import "github.com/caos/orbiter/internal/operator/orbiter"

type DesiredV0 struct {
	Common *orbiter.Common `yaml:",inline"`
	Deps   map[string]*orbiter.Tree
}

type SecretsV0 struct {
	Common *orbiter.Common `yaml:",inline"`
	Deps   map[string]*orbiter.Tree
}

type Current struct {
	Common *orbiter.Common `yaml:",inline"`
	Deps   map[string]*orbiter.Tree
}
