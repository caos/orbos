package v1beta2

import "github.com/caos/orbos/internal/operator/boom/api/v1beta2/storage"

type LogCollection struct {
	Deploy     bool          `json:"deploy" yaml:"deploy"`
	FluentdPVC *storage.Spec `json:"fluentdStorage,omitempty" yaml:"fluentdStorage,omitempty"`
}
