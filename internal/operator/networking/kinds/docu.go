package kinds

import (
	"github.com/caos/orbos/internal/docu"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking"
	"github.com/caos/orbos/internal/operator/networking/kinds/orb"
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

	infos = append(infos, networking.GetDocuInfo()...)
	return infos
}
