package orb

import (
	"github.com/caos/orbos/internal/secret"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type Orb struct {
	Path      string `yaml:"-"`
	URL       string
	Repokey   string
	Masterkey string
}

func ParseOrbConfig(orbConfigPath string) (*Orb, error) {
	gitOrbConfig, err := ioutil.ReadFile(orbConfigPath)

	if err != nil {
		return nil, errors.Wrap(err, "unable to read orbconfig")
	}

	orb := &Orb{}
	if err := yaml.Unmarshal(gitOrbConfig, orb); err != nil {
		return nil, errors.Wrap(err, "unable to parse orbconfig")
	}

	orb.Path = orbConfigPath
	secret.Masterkey = orb.Masterkey
	return orb, nil
}

func (o *Orb) WriteBackOrbConfig() error {
	data, err := yaml.Marshal(o)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(o.Path, data, os.ModePerm)
}
