package orb

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caos/orbos/internal/helpers"

	"github.com/caos/orbos/internal/secret"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Orb struct {
	Path      string `yaml:"-"`
	URL       string
	Repokey   string
	Masterkey string
}

func (o *Orb) IsConnectable() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("repository is not connectable: %w", err)
		}
	}()
	if o.URL == "" {
		err = helpers.Concat(err, errors.New("repository url is missing"))
	}

	if o.Repokey == "" {
		err = helpers.Concat(err, errors.New("repository key is missing"))
	}
	return err
}

func (o *Orb) IsComplete() (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("orbconfig is incomplete: %w", err)
		}
	}()

	if o.Masterkey == "" {
		err = helpers.Concat(err, errors.New("master key is missing"))
	}

	if o.Path == "" {
		err = helpers.Concat(err, errors.New("file path is missing"))
	}

	return helpers.Concat(err, o.IsConnectable())
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
