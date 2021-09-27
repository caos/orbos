package static

import (
	"github.com/caos/orbos/v5/pkg/secret"
)

func getSecretsMap(desiredKind *DesiredV0) map[string]*secret.Secret {

	if desiredKind.Spec.Keys == nil {
		desiredKind.Spec.Keys = &Keys{}
	}

	if desiredKind.Spec.Keys.BootstrapKeyPrivate == nil {
		desiredKind.Spec.Keys.BootstrapKeyPrivate = &secret.Secret{}
	}

	if desiredKind.Spec.Keys.BootstrapKeyPublic == nil {
		desiredKind.Spec.Keys.BootstrapKeyPublic = &secret.Secret{}
	}

	if desiredKind.Spec.Keys.MaintenanceKeyPrivate == nil {
		desiredKind.Spec.Keys.MaintenanceKeyPrivate = &secret.Secret{}
	}

	if desiredKind.Spec.Keys.MaintenanceKeyPublic == nil {
		desiredKind.Spec.Keys.MaintenanceKeyPublic = &secret.Secret{}
	}

	return map[string]*secret.Secret{
		"bootstrapkeyprivate":   desiredKind.Spec.Keys.BootstrapKeyPrivate,
		"bootstrapkeypublic":    desiredKind.Spec.Keys.BootstrapKeyPublic,
		"maintenancekeyprivate": desiredKind.Spec.Keys.MaintenanceKeyPrivate,
		"maintenancekeypublic":  desiredKind.Spec.Keys.MaintenanceKeyPublic,
	}
}
