package repository

import (
	"strings"

	"github.com/caos/orbos/pkg/secret/read"

	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/info"
	"github.com/caos/orbos/internal/operator/boom/application/resources"
	"github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/mntr"
)

type Repository struct {
	URL                 string
	UsernameSecret      *secret `yaml:"usernameSecret,omitempty"`
	PasswordSecret      *secret `yaml:"passwordSecret,omitempty"`
	SSHPrivateKeySecret *secret `yaml:"sshPrivateKeySecret,omitempty"`
}

const (
	cert = "certificate"
	user = "username"
	pw   = "password"
)

type secret struct {
	Name string
	Key  string
}

func getSecretName(name string, ty string) string {
	return strings.Join([]string{info.GetName().String(), "repo", name, ty}, "-")
}

func getSecretKey(ty string) string {
	return ty
}

func GetSecrets(spec *reconciling.Reconciling) []interface{} {
	secrets := make([]interface{}, 0)
	namespace := "caos-system"

	for _, v := range spec.Repositories {
		if read.IsCrdSecret(v.Username, v.ExistingUsernameSecret) {

			data := map[string]string{
				getSecretKey(user): v.Username.Value,
			}

			conf := &resources.SecretConfig{
				Name:      getSecretName(v.Name, user),
				Namespace: namespace,
				Labels:    labels.GetAllApplicationLabels(info.GetName()),
				Data:      data,
			}
			secretRes := resources.NewSecret(conf)
			secrets = append(secrets, secretRes)
		}
		if read.IsCrdSecret(v.Password, v.ExistingPasswordSecret) {

			data := map[string]string{
				getSecretKey(pw): v.Password.Value,
			}

			conf := &resources.SecretConfig{
				Name:      getSecretName(v.Name, pw),
				Namespace: namespace,
				Labels:    labels.GetAllApplicationLabels(info.GetName()),
				Data:      data,
			}
			secretRes := resources.NewSecret(conf)
			secrets = append(secrets, secretRes)
		}
		if read.IsCrdSecret(v.Certificate, v.ExistingCertificateSecret) {
			data := map[string]string{
				getSecretKey(cert): v.Certificate.Value,
			}

			conf := &resources.SecretConfig{
				Name:      getSecretName(v.Name, cert),
				Namespace: namespace,
				Labels:    labels.GetAllApplicationLabels(info.GetName()),
				Data:      data,
			}
			secretRes := resources.NewSecret(conf)
			secrets = append(secrets, secretRes)
		}
	}

	return secrets
}

func GetFromSpec(monitor mntr.Monitor, spec *reconciling.Reconciling) []*Repository {
	repositories := make([]*Repository, 0)

	if spec.Repositories == nil || len(spec.Repositories) == 0 {
		return repositories
	}

	for _, v := range spec.Repositories {
		var us, ps, ssh *secret
		if read.IsCrdSecret(v.Username, v.ExistingUsernameSecret) {
			us = &secret{
				Name: getSecretName(v.Name, user),
				Key:  getSecretKey(user),
			}
		} else if read.IsExistentSecret(v.Username, v.ExistingUsernameSecret) {
			us = &secret{
				Name: v.ExistingUsernameSecret.Name,
				Key:  v.ExistingUsernameSecret.Key,
			}
		}

		if read.IsCrdSecret(v.Password, v.ExistingPasswordSecret) {
			ps = &secret{
				Name: getSecretName(v.Name, pw),
				Key:  getSecretKey(pw),
			}
		} else if read.IsExistentSecret(v.Password, v.ExistingPasswordSecret) {
			ps = &secret{
				Name: v.ExistingPasswordSecret.Name,
				Key:  v.ExistingPasswordSecret.Key,
			}
		}

		if read.IsCrdSecret(v.Certificate, v.ExistingCertificateSecret) {
			ssh = &secret{
				Name: getSecretName(v.Name, cert),
				Key:  getSecretKey(cert),
			}
		} else if read.IsExistentSecret(v.Certificate, v.ExistingCertificateSecret) {
			ssh = &secret{
				Name: v.ExistingCertificateSecret.Name,
				Key:  v.ExistingCertificateSecret.Key,
			}
		}

		repo := &Repository{
			URL:                 v.URL,
			UsernameSecret:      us,
			PasswordSecret:      ps,
			SSHPrivateKeySecret: ssh,
		}
		repositories = append(repositories, repo)
	}

	return repositories
}
