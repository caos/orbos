package repository

import (
	"github.com/caos/orbos/internal/secret"
)

// Repository: For a repository there are two types, with ssh-connection where an url and a certificate have to be provided and an https-connection where an URL, username and password have to be provided.
type Repository struct {
	//Internal used name
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Prefix where the credential should be used (starting "git@" or "https://" )
	URL      string         `json:"url,omitempty" yaml:"url,omitempty"`
	Username *secret.Secret `yaml:"username,omitempty"`
	//Existing secret used for username
	ExistingUsernameSecret *secret.Existing `json:"existingUsernameSecret,omitempty" yaml:"existingUsernameSecret,omitempty"`
	Password               *secret.Secret   `yaml:"password,omitempty"`
	//Existing secret used for password
	ExistingPasswordSecret *secret.Existing `json:"existingPasswordSecret,omitempty" yaml:"existingPasswordSecret,omitempty"`
	Certificate            *secret.Secret   `yaml:"certificate,omitempty"`
	//Existing secret used for certificate
	ExistingCertificateSecret *secret.Existing `json:"existingCertificateSecret,omitempty" yaml:"existingCertificateSecret,omitempty"`
}
