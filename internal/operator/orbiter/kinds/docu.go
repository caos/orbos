package kinds

import (
	"github.com/caos/orbos/internal/docu"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers"
)

func GetDocuInfo() []*docu.Type {
	path, orbVersions := orb.GetDocuInfo()

	infos := []*docu.Type{{
		Name: "orb",
		Kinds: []*docu.Info{
			{
				Path:     path,
				Kind:     "orbiter.caos.ch/Orb",
				Versions: orbVersions,
			},
		},
	}}

	infos = append(infos, providers.GetDocuInfo()...)
	infos = append(infos, clusters.GetDocuInfo()...)
	return infos
}
