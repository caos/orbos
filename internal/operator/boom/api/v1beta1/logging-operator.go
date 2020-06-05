package v1beta1

import "github.com/caos/orbos/internal/operator/boom/api/v1beta1/storage"

type LoggingOperator struct {
	Deploy     bool          `json:"deploy" yaml:"deploy"`
	FluentdPVC *storage.Spec `json:"fluentdStorage,omitempty" yaml:"fluentdStorage,omitempty"`
}
