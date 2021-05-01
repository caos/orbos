package admin

import (
	"github.com/caos/orbos/pkg/secret"
)

// Admin: Not defining the admin credentials results in an user admin with password admin.
type Admin struct {
	Username *secret.Secret `json:"username,omitempty" yaml:"username,omitempty"`
	Password *secret.Secret `json:"password,omitempty" yaml:"password,omitempty"`
	//Existing Secret containing username and password
	ExistingUsername *secret.Existing `json:"existingUsername,omitempty" yaml:"existingUsername,omitempty"`
	ExistingPassword *secret.Existing `json:"existingPassword,omitempty" yaml:"existingPassword,omitempty"`
}

func (a *Admin) IsZero() bool {
	if (a.Username == nil || a.Username.IsZero()) &&
		(a.Password == nil || a.Password.IsZero()) &&
		(a.ExistingUsername == nil || a.ExistingUsername.IsZero()) &&
		(a.ExistingPassword == nil || a.ExistingPassword.IsZero()) {
		return true
	}
	return false
}

func (a *Admin) InitSecrets() {
	if a.Username == nil {
		a.Username = &secret.Secret{}
	}
	if a.ExistingUsername == nil {
		a.ExistingUsername = &secret.Existing{}
	}
	if a.Password == nil {
		a.Password = &secret.Secret{}
	}
	if a.ExistingPassword == nil {
		a.ExistingPassword = &secret.Existing{}
	}
}
