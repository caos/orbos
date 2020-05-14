package storage

type Spec struct {
	StorageClass string   `json:"storageClass,omitempty" yaml:"storageClass,omitempty"`
	AccessModes  []string `json:"accessModes,omitempty" yaml:"accessModes,omitempty"`
	Size         string   `json:"size,omitempty" yaml:"size,omitempty"`
}
