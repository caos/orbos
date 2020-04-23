package v1beta1

import "github.com/caos/orbiter/internal/operator/boom/api/v1beta1/storage"

type LoggingOperator struct {
	Deploy     bool          `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	FluentdPVC *storage.Spec `json:"fluentdStorage,omitempty" yaml:"fluentdStorage,omitempty"`
}
