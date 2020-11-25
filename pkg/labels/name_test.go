package labels_test

import (
	"testing"

	"github.com/caos/orbos/pkg/labels"
	"gopkg.in/yaml.v3"
)

func expectValidNameLabels(t *testing.T, name string) *labels.Name {
	l, err := labels.ForName(validComponentLabels(t), name)
	if err != nil {
		t.Fatal()
	}
	return l
}

func validNameLabels(t *testing.T) *labels.Name {
	return expectValidNameLabels(t, "testcase")
}

func TestNameLabels_Equal(t *testing.T) {
	expectValueEquality(
		t,
		validNameLabels(t),
		validNameLabels(t),
		expectValidNameLabels(t, "testcase"))
}

func TestNameLabels_MarshalYAML(t *testing.T) {
	marshalled, err := yaml.Marshal(validNameLabels(t))
	if err != nil {
		t.Error(err, "expected successful mashalling")
	}

	expected := `app.kubernetes.io/name: testcase
app.kubernetes.io/component: testSuite
orbos.ch/kind: testing.caos.ch/TestSuite
orbos.ch/apiversion: v1
app.kubernetes.io/version: testing-dev
app.kubernetes.io/part-of: ORBOS
app.kubernetes.io/managed-by: TEST_OPERATOR_LABELS
`
	if string(marshalled) != expected {
		t.Errorf("expected \n%s\n but got \n%s\n", expected, string(marshalled))
	}
}
