package admin

import (
	"github.com/caos/orbos/internal/secret"
)

// Admin: Not defining the admin credentials results in an user admin with password admin.
type Admin struct {
	Username *secret.Secret `json:"username,omitempty" yaml:"username,omitempty"`
	Password *secret.Secret `yaml:"password,omitempty"`
	//Existing Secret containing username and password
	ExistingSecret *secret.ExistingIDSecret `json:"existingSecret,omitempty" yaml:"existingSecret,omitempty"`
}
