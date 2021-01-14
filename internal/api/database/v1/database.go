// +kubebuilder:object:generate=true
// +groupName=caos.ch
// +kubebuilder:storageversion
package v1

import (
	orbdb "github.com/caos/orbos/internal/operator/database/kinds/orb"
	"github.com/caos/orbos/pkg/tree"
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

// +kubebuilder:object:root=true
// +kubebuilder:crd=Database
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Spec   `json:"spec,omitempty"`
	Status Status `json:"status,omitempty"`
}

type Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

type Spec struct {
	Common   *tree.Common `json:",inline" yaml:",inline"`
	Spec     *orbdb.Spec  `json:"spec" yaml:"spec"`
	Database *Empty       `json:"database" yaml:"database"`
}

type Empty struct{}

func init() {
	SchemeBuilder.Register(&Database{})
}
