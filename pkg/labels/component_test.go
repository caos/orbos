package labels_test

import (
	"testing"

	"github.com/caos/orbos/v5/pkg/labels"
)

func expectValidComponentLabels(t *testing.T, component string) *labels.Component {
	l, err := labels.ForComponent(validAPILabels(t), component)
	if err != nil {
		t.Fatal()
	}
	return l
}

func validComponentLabels(t *testing.T) *labels.Component {
	return expectValidComponentLabels(t, "testSuite")
}

func TestComponentLabels_Equal(t *testing.T) {
	expectValueEquality(
		t,
		validComponentLabels(t),
		validComponentLabels(t),
		expectValidComponentLabels(t, "somethingElse"))
}

func TestComponentLabels_MarshalYAML(t *testing.T) {
	expectNotMarshallable(t, validComponentLabels(t))
}
