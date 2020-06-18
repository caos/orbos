package repository

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/info"
	"github.com/caos/orbos/internal/operator/boom/application/resources"
	"github.com/caos/orbos/internal/operator/boom/labels"
	helper2 "github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"
	"strings"
)

type Repository struct {
	URL                 string
	UsernameSecret      *secret `yaml:"usernameSecret,omitempty"`
	PasswordSecret      *secret `yaml:"passwordSecret,omitempty"`
	SSHPrivateKeySecret *secret `yaml:"sshPrivateKeySecret,omitempty"`
}

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

func GetSecrets(spec *argocd.Argocd) []interface{} {
	secrets := make([]interface{}, 0)
	namespace := "caos-system"

	for _, v := range spec.Repositories {
		if helper2.IsCrdSecret(v.Username, v.ExistingUsernameSecret) {
			ty := "username"

			data := map[string]string{
				getSecretKey(ty): v.Username.Value,
			}

			conf := &resources.SecretConfig{
				Name:      getSecretName(v.Name, ty),
				Namespace: namespace,
				Labels:    labels.GetAllApplicationLabels(info.GetName()),
				Data:      data,
			}
			secretRes := resources.NewSecret(conf)
			secrets = append(secrets, secretRes)
		}
		if helper2.IsCrdSecret(v.Password, v.ExistingPasswordSecret) {
			ty := "password"

			data := map[string]string{
				getSecretKey(ty): v.Password.Value,
			}

			conf := &resources.SecretConfig{
				Name:      getSecretName(v.Name, ty),
				Namespace: namespace,
				Labels:    labels.GetAllApplicationLabels(info.GetName()),
				Data:      data,
			}
			secretRes := resources.NewSecret(conf)
			secrets = append(secrets, secretRes)
		}
		if helper2.IsCrdSecret(v.Certificate, v.ExistingCertificateSecret) {
			ty := "certificate"

			data := map[string]string{
				getSecretKey(ty): v.Certificate.Value,
			}

			conf := &resources.SecretConfig{
				Name:      getSecretName(v.Name, ty),
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

func GetFromSpec(monitor mntr.Monitor, spec *argocd.Argocd) []*Repository {
	repositories := make([]*Repository, 0)

	if spec.Repositories == nil || len(spec.Repositories) == 0 {
		return repositories
	}

	for _, v := range spec.Repositories {
		var us, ps, ssh *secret
		if helper2.IsCrdSecret(v.Username, v.ExistingUsernameSecret) {
			ty := "username"
			us = &secret{
				Name: getSecretName(v.Name, ty),
				Key:  getSecretKey(ty),
			}
		} else if helper2.IsExistentSecret(v.Username, v.ExistingUsernameSecret) {
			us = &secret{
				Name: v.ExistingUsernameSecret.Name,
				Key:  v.ExistingUsernameSecret.Key,
			}
		}

		if helper2.IsCrdSecret(v.Password, v.ExistingPasswordSecret) {
			ty := "password"
			ps = &secret{
				Name: getSecretName(v.Name, ty),
				Key:  getSecretKey(ty),
			}
		} else if helper2.IsExistentSecret(v.Password, v.ExistingPasswordSecret) {
			ps = &secret{
				Name: v.ExistingPasswordSecret.Name,
				Key:  v.ExistingPasswordSecret.Key,
			}
		}

		if helper2.IsCrdSecret(v.Certificate, v.ExistingCertificateSecret) {
			ty := "username"
			ssh = &secret{
				Name: getSecretName(v.Name, ty),
				Key:  getSecretKey(ty),
			}
		} else if helper2.IsExistentSecret(v.Certificate, v.ExistingCertificateSecret) {
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
