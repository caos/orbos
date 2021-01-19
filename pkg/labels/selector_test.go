package labels_test

import (
	"testing"

	"github.com/caos/orbos/pkg/labels"
	"gopkg.in/yaml.v3"
)

func TestSelectorLabels_Equal(t *testing.T) {
	nameLabels := validNameLabels(t)
	expectValueEquality(
		t,
		labels.DeriveSelector(nameLabels, true),
		labels.DeriveSelector(nameLabels, true),
		labels.DeriveSelector(nameLabels, false),
	)
}

func TestSelectorLabels_Open_MarshalYAML(t *testing.T) {
	testSelectorLabels(t, labels.DeriveSelector(validNameLabels(t), true), `orbos.ch/select: true
app.kubernetes.io/component: testSuite
`)
}

func TestSelectorLabels_Close_MarshalYAML(t *testing.T) {
	testSelectorLabels(t, labels.DeriveSelector(validNameLabels(t), false), `orbos.ch/select: true
app.kubernetes.io/name: testcase
app.kubernetes.io/component: testSuite
app.kubernetes.io/managed-by: TEST_OPERATOR_LABELS
app.kubernetes.io/part-of: ORBOS
`)
}

func testSelectorLabels(t *testing.T, selector *labels.Selector, expected string) {
	marshalled, err := yaml.Marshal(selector)
	if err != nil {
		t.Error("expected successful mashalling")
	}

	if string(marshalled) != expected {
		t.Errorf("expected \n%s\n but got \n%s\n", expected, string(marshalled))
	}
}
