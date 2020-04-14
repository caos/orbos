package helpers_test

import (
	"github.com/caos/orbiter/internal/helpers"
	"testing"
)

func TestRandomStringRunes(t *testing.T){
	expected := 5
	str := helpers.RandomString(expected)
	actual := len(str)
	if actual != expected {
		t.Errorf("Expected length %d, actual length %d", expected, actual)
	}
}
