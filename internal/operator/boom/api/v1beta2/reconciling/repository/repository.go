package repository

import (
	"github.com/caos/orbos/internal/secret"
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
