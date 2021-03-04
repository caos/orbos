// +kubebuilder:object:generate=true
// +groupName=caos.ch
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "caos.ch", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:crd=BOOM
type BOOM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *Empty `json:"spec,omitempty"`
	Status Status `json:"status,omitempty"`
}

type Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

type Empty struct{}

// +kubebuilder:object:root=true
type BOOMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BOOM `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BOOM{}, &BOOMList{})
}

func GetEmpty(namespace, name string) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "caos.ch/v1",
		"kind":       "BOOM",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"forceApply": true,
			"postApply": map[string]interface{}{
				"deploy": false,
			},
			"preApply": map[string]interface{}{
				"deploy": false,
			},
			"metricCollection": map[string]interface{}{
				"deploy": false,
			},
			"logCollection": map[string]interface{}{
				"deploy": false,
			},
			"nodeMetricsExporter": map[string]interface{}{
				"deploy": false,
			},
			"systemdMetricsExporter": map[string]interface{}{
				"deploy": false,
			},
			"monitoring": map[string]interface{}{
				"deploy": false,
			},
			"apiGateway": map[string]interface{}{
				"deploy": false,
			},
			"kubeMetricsExporter": map[string]interface{}{
				"deploy": false,
			},
			"reconciling": map[string]interface{}{
				"deploy": false,
			},
			"metricsPersisting": map[string]interface{}{
				"deploy": false,
			},
			"logsPersisting": map[string]interface{}{
				"deploy": false,
			},
			"metricsServer": map[string]interface{}{
				"deploy": true,
			},
		},
	}
}
