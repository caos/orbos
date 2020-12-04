package labels_test

import (
	"testing"

	"github.com/caos/orbos/pkg/labels"
	"gopkg.in/yaml.v3"
)

func validSelectableLabels(t *testing.T) *labels.Selectable {
	return labels.AsSelectable(validNameLabels(t))
}

func TestSelectableLabels_MarshalYAML(t *testing.T) {
	marshalled, err := yaml.Marshal(validSelectableLabels(t))
	if err != nil {
		t.Error(err, "expected successful mashalling")
	}

	expected := "orbos.ch/select: true\n" + validLabels
	if string(marshalled) != expected {
		t.Errorf("expected \n%s\n but got \n%s\n", expected, string(marshalled))
	}
}
