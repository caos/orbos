package bucket

import (
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	//Configuration for backups in GCS bucket
	Spec *Spec
}

type Spec struct {
	//Verbose flag to set debug-level to debug
	Verbose bool
	//Cron-job interval when the backups should be done
	Cron string `yaml:"cron,omitempty"`
	//Name of the bucket in the google cloud project as the service account exists
	Bucket string `yaml:"bucket,omitempty"`
	//JSON for the service account used to write the backup to the bucket
	ServiceAccountJSON *secret.Secret `yaml:"serviceAccountJSON,omitempty"`
}

func (s *Spec) IsZero() bool {
	if (s.ServiceAccountJSON == nil || s.ServiceAccountJSON.IsZero()) &&
		!s.Verbose &&
		s.Cron == "" &&
		s.Bucket == "" {
		return true
	}
	return false
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: desiredTree.Common,
		Spec:   &Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
