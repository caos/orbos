package labels_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/pkg/labels"
)

func expectValidOperatorLabels(t *testing.T, operator, version string) *labels.Operator {
	l, err := labels.ForOperator(operator, version)
	if err != nil {
		t.Fatal()
	}
	return l
}

func validOperatorLabels(t *testing.T) *labels.Operator {
	return expectValidOperatorLabels(t, "TEST_OPERATOR_LABELS", "testing-dev")
}

func TestOperatorLabels_Equal(t *testing.T) {
	expectValueEquality(
		t,
		validOperatorLabels(t),
		validOperatorLabels(t),
		expectValidOperatorLabels(t, "TWO", "testing-dev"),
	)
}

func TestOperatorLabels_MarshalYAML(t *testing.T) {
	testCase := validOperatorLabels(t)
	_, err := yaml.Marshal(testCase)
	if err == nil {
		t.Error("expected full set of labels")
	}
}
