package kubernetes

import (
	"github.com/caos/orbos/pkg/labels"

	"github.com/caos/orbos/internal/orb"
	"gopkg.in/yaml.v2"
	core "k8s.io/api/core/v1"
	mach "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caos/orbos/mntr"
)

func EnsureCommonArtifacts(monitor mntr.Monitor, client *Client) error {

	monitor.Debug("Ensuring common artifacts")

	return client.ApplyNamespace(&core.Namespace{
		ObjectMeta: mach.ObjectMeta{
			Name: "caos-system",
			Labels: map[string]string{
				"name":                      "caos-system",
				"app.kubernetes.io/part-of": "orbos",
			},
		},
	})
}

func EnsureOrbconfigSecret(monitor mntr.Monitor, client *Client, orb *orb.Orb) error {
	monitor.Debug("Ensuring configuration artifacts")

	orbfile, err := yaml.Marshal(orb)
	if err != nil {
		return err
	}

	if err := client.ApplySecret(&core.Secret{
		ObjectMeta: mach.ObjectMeta{
			Name:      "caos",
			Namespace: "caos-system",
			Labels: map[string]string{
				"app.kubernetes.io/part-of": "orbos",
			},
		},
		StringData: map[string]string{
			"orbconfig": string(orbfile),
		},
	}); err != nil {
		return err
	}

	return nil
}

func toNameLabels(apiLabels *labels.API, operatorName string) *labels.Name {
	return labels.MustForName(labels.MustForComponent(apiLabels, "operator"), operatorName)
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
