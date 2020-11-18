package storage

import (
	storagev1beta2 "github.com/caos/orbos/internal/operator/boom/api/latest/storage"
	storagev1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1/storage"
)

func V1beta1Tov1beta2(old *storagev1beta1.Spec) *storagev1beta2.Spec {
	if old == nil {
		return nil
	}

	ret := &storagev1beta2.Spec{
		StorageClass: old.StorageClass,
		Size:         old.Size,
	}
	if old.AccessModes != nil && len(old.AccessModes) > 0 {
		for _, v := range old.AccessModes {
			ret.AccessModes = append(ret.AccessModes, v)
		}
	}

	return ret
}
