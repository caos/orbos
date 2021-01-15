package labels_test

import (
	"testing"

	"github.com/caos/orbos/pkg/labels"
)

func expectValidAPILabels(t *testing.T, kind, version string) *labels.API {
	l, err := labels.ForAPI(validOperatorLabels(t), kind, version)
	if err != nil {
		t.Fatal()
	}
	return l
}

func validAPILabels(t *testing.T) *labels.API {
	return expectValidAPILabels(t, "testing.caos.ch/TestSuite", "v1")
}

func TestAPILabels_Equal(t *testing.T) {
	expectValueEquality(
		t,
		validAPILabels(t),
		validAPILabels(t),
		expectValidAPILabels(t, "testing.caos.ch/TestSuite", "v2"))
}

func TestAPILabels_MarshalYAML(t *testing.T) {
	expectNotMarshallable(t, validAPILabels(t))
}
