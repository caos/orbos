package cs

import (
	"github.com/caos/orbos/internal/secret"
)

func getSecretsMap(desiredKind *Desired) map[string]*secret.Secret {
	if desiredKind.Spec.APIToken == nil {
		desiredKind.Spec.APIToken = &secret.Secret{}
	}

	if desiredKind.Spec.SSHKey == nil {
		desiredKind.Spec.SSHKey = &SSHKey{}
	}

	if desiredKind.Spec.SSHKey.Public == nil {
		desiredKind.Spec.SSHKey.Public = &secret.Secret{}
	}

	if desiredKind.Spec.SSHKey.Private == nil {
		desiredKind.Spec.SSHKey.Private = &secret.Secret{}
	}

	return map[string]*secret.Secret{
		"apitoken":          desiredKind.Spec.APIToken,
		"sshkeyprivate":     desiredKind.Spec.SSHKey.Private,
		"sshkeypublic":      desiredKind.Spec.SSHKey.Public,
		"rootsshkeyprivate": desiredKind.Spec.RootSSHKey.Private,
		"rootsshkeypublic":  desiredKind.Spec.RootSSHKey.Public,
	}
}
