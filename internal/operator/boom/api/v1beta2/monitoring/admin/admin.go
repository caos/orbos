package admin

import (
	secret2 "github.com/caos/orbos/pkg/secret"
)

// Admin: Not defining the admin credentials results in an user admin with password admin.
type Admin struct {
	Username *secret2.Secret `json:"username,omitempty" yaml:"username,omitempty"`
	Password *secret2.Secret `yaml:"password,omitempty"`
	//Existing Secret containing username and password
	ExistingSecret *secret2.ExistingIDSecret `json:"existingSecret,omitempty" yaml:"existingSecret,omitempty"`
}
