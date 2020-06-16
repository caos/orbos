package admin

import (
	"github.com/caos/orbos/internal/secret"
)

type Admin struct {
	Username       *secret.Secret           `json:"username,omitempty" yaml:"username,omitempty"`
	Password       *secret.Secret           `yaml:"password,omitempty"`
	ExistingSecret *secret.ExistingIDSecret `json:"existingSecret,omitempty" yaml:"existingSecret,omitempty"`
}
