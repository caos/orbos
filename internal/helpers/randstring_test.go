package helpers_test

import (
	"testing"

	"github.com/caos/orbos/v5/internal/helpers"
)

func TestRandomStringRunes(t *testing.T) {
	expected := 5
	str := helpers.RandomString(expected)
	actual := len(str)
	if actual != expected {
		t.Errorf("Expected length %d, actual length %d", expected, actual)
	}
}
