package network

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/network"
	"github.com/caos/orbos/internal/operator/boom/application/applications/apigateway/crds"
)

func GetHostConfig(spec *network.Network) *crds.HostConfig {
	return &crds.HostConfig{
		Name:             spec.Domain,
		Namespace:        "caos-system",
		InsecureAction:   "Redirect",
		Hostname:         spec.Domain,
		AcmeProvider:     spec.AcmeAuthority,
		PrivateKeySecret: spec.Domain,
		Email:            spec.Email,
		TLSSecret:        spec.Domain,
	}
}

func GetMappingConfig(spec *network.Network) *crds.MappingConfig {
	return &crds.MappingConfig{
		Name:      "argocd",
		Namespace: "caos-system",
		Prefix:    "/",
		Service:   "https://argocd-server.caos-system:443",
		Host:      spec.Domain,
	}
}
