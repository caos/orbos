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

	expected := `orbos.ch/select: true
app.kubernetes.io/name: testcase
app.kubernetes.io/component: testSuite
orbos.ch/kind: testing.caos.ch/TestSuite
orbos.ch/apiversion: v1
app.kubernetes.io/managed-by: TEST_OPERATOR_LABELS
app.kubernetes.io/version: testing-dev
app.kubernetes.io/part-of: ORBOS
`
	if string(marshalled) != expected {
		t.Errorf("expected \n%s\n but got \n%s\n", expected, string(marshalled))
	}
}
