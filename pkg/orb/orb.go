package orb

import "github.com/caos/orbos/internal/orb"

type Orb orb.Orb

func ParseOrbConfig(orbConfigPath string) (*Orb, error) {
	orbConfig, err := orb.ParseOrbConfig(orbConfigPath)
	if err != nil {
		return nil, err
	}
	return &Orb{
		Repokey:   orbConfig.Repokey,
		Masterkey: orbConfig.Masterkey,
		URL:       orbConfig.URL,
		Path:      orbConfig.Path,
	}, nil
}
