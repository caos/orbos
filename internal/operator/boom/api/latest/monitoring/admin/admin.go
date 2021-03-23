package admin

import (
	"github.com/caos/orbos/pkg/secret"
)

// Admin: Not defining the admin credentials results in an user admin with password admin.
type Admin struct {
	Username *secret.Secret `json:"username,omitempty" yaml:"username,omitempty"`
	Password *secret.Secret `json:"password,omitempty" yaml:"password,omitempty"`
	//Existing Secret containing username and password
	ExistingSecret *secret.ExistingIDSecret `json:"existingSecret,omitempty" yaml:"existingSecret,omitempty"`
}

func (a *Admin) IsZero() bool {
	if (a.Username == nil || a.Username.IsZero()) &&
		(a.Username == nil || a.Username.IsZero()) &&
		a.ExistingSecret == nil {
		return true
	}
	return false
}

func (a *Admin) InitSecrets() {
	if a.Username == nil {
		a.Username = &secret.Secret{}
	}
	if a.ExistingSecret == nil {
		a.ExistingSecret = &secret.ExistingIDSecret{}
	}
	if a.Password == nil {
		a.Password = &secret.Secret{}
	}
}
