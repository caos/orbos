package v1beta2

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/storage"
	"github.com/caos/orbos/pkg/kubernetes/k8s"
)

type LogsPersisting struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Spec to define which logs will get persisted
	//@default: nil
	Logs *Logs `json:"logs,omitempty" yaml:"logs,omitempty"`
	//Spec to define how the persistence should be handled
	//@default: nil
	Storage *storage.Spec `json:"storage,omitempty" yaml:"storage,omitempty"`
	//Flag if loki-output should be a clusteroutput instead a output crd
	//@default: false
	ClusterOutput bool `json:"clusterOutput,omitempty" yaml:"clusterOutput,omitempty"`
	//NodeSelector for statefulset
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run loki on nodes
	Tolerations k8s.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// Logs: When the logs spec is nil all logs will get persisted in loki.
type Logs struct {
	//Bool if logs will get persisted for ambassador
	Ambassador bool `json:"ambassador"`
	//Bool if logs will get persisted for grafana
	Grafana bool `json:"grafana"`
	//Bool if logs will get persisted for argo-cd
	Argocd bool `json:"argocd"`
	//Bool if logs will get persisted for kube-state-metrics
	KubeStateMetrics bool `json:"kube-state-metrics" yaml:"kube-state-metrics"`
	//Bool if logs will get persisted for prometheus-node-exporter
	PrometheusNodeExporter bool `json:"prometheus-node-exporter"  yaml:"prometheus-node-exporter"`
	//Bool if logs will get persisted for prometheus-operator
	PrometheusOperator bool `json:"prometheus-operator" yaml:"prometheus-operator"`
	//Bool if logs will get persisted for logging-operator
	LoggingOperator bool `json:"logging-operator" yaml:"logging-operator"`
	//Bool if logs will get persisted for loki
	Loki bool `json:"loki"`
	//Bool if logs will get persisted for prometheus
	Prometheus bool `json:"prometheus"`
}
