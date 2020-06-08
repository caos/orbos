package repository

import (
	"github.com/caos/orbos/internal/secret"
	"reflect"
)

type Repository struct {
	Name                      string           `json:"name,omitempty" yaml:"name,omitempty"`
	URL                       string           `json:"url,omitempty" yaml:"url,omitempty"`
	Username                  *secret.Secret   `yaml:"username,omitempty"`
	ExistingUsernameSecret    *secret.Existing `json:"existingUsernameSecret,omitempty" yaml:"existingUsernameSecret,omitempty"`
	Password                  *secret.Secret   `yaml:"password,omitempty"`
	ExistingPasswordSecret    *secret.Existing `json:"existingPasswordSecret,omitempty" yaml:"existingPasswordSecret,omitempty"`
	Certificate               *secret.Secret   `yaml:"certificate,omitempty"`
	ExistingCertificateSecret *secret.Existing `json:"existingCertificateSecret,omitempty" yaml:"existingCertificateSecret,omitempty"`
}

func ClearEmpty(x *Repository) *Repository {
	if x == nil {
		return nil
	}

	marshaled := Repository{
		Name:                      x.Name,
		URL:                       x.URL,
		Username:                  secret.ClearEmpty(x.Username),
		ExistingUsernameSecret:    x.ExistingUsernameSecret,
		Password:                  secret.ClearEmpty(x.Password),
		ExistingPasswordSecret:    x.ExistingPasswordSecret,
		Certificate:               secret.ClearEmpty(x.Certificate),
		ExistingCertificateSecret: x.ExistingCertificateSecret,
	}

	if reflect.DeepEqual(marshaled, Repository{}) {
		return nil
	}
	return &marshaled
}
