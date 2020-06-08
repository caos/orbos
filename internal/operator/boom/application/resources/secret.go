package resources

import "encoding/base64"

type SecretConfig struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Data      map[string]string
}

type Secret struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   *Metadata         `yaml:"metadata"`
	Data       map[string]string `yaml:"data"`
	Type       string            `yaml:"type"`
}

func NewSecret(conf *SecretConfig) *Secret {
	encodedData := make(map[string]string, 0)
	for k, v := range conf.Data {
		encodedData[k] = base64.StdEncoding.EncodeToString([]byte(v))
	}

	return &Secret{
		APIVersion: "v1",
		Kind:       "Secret",
		Metadata: &Metadata{
			Name:      conf.Name,
			Namespace: conf.Namespace,
			Labels:    conf.Labels,
		},
		Type: "Opaque",
		Data: encodedData,
	}
}
