package labels_test

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func expectNotMarshallable(t *testing.T, labels interface{}) {
	_, err := yaml.Marshal(labels)
	if err == nil {
		t.Error("expected full set of labels")
	}
}

func expectValueEquality(t *testing.T, one labels.comparable, oneTick labels.comparable, two labels.comparable) {

	if one.Equal(two) {
		t.Error("Expected unequal")
	}

	if !one.Equal(oneTick) {
		t.Error("Expected value equality")
	}
}
