package network

import (
	networkv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1/network"
	networkv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2/network"
)

func V1beta1Tov1beta2(old *networkv1beta1.Network) *networkv1beta2.Network {
	if old == nil {
		return nil
	}

	return &networkv1beta2.Network{
		Domain:        old.Domain,
		Email:         old.Email,
		AcmeAuthority: old.AcmeAuthority,
	}
}
