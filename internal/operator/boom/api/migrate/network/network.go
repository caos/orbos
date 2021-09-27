package network

import (
	networkvv1beta2 "github.com/caos/orbos/v5/internal/operator/boom/api/latest/network"
	networkv1beta1 "github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/network"
)

func V1beta1Tov1beta2(old *networkv1beta1.Network) *networkvv1beta2.Network {
	if old == nil {
		return nil
	}

	return &networkvv1beta2.Network{
		Domain:        old.Domain,
		Email:         old.Email,
		AcmeAuthority: old.AcmeAuthority,
	}
}
