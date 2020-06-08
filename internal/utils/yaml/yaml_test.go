package yaml

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type teststruct struct {
	Test string `yaml:"test"`
}

func TestHelper_YamlToString(t *testing.T) {
	var str string
	str, err := New("testfiles/struct.yaml").ToString()
	assert.NoError(t, err)

	assert.Equal(t, "test: test", str)
}

func TestHelper_YamlToString_nonexistent(t *testing.T) {
	var str string
	str, err := New("testfiles/nonexistent.yaml").ToString()
	assert.Error(t, err)
	assert.Empty(t, str)
}

func TestHelper_YamlToStruct(t *testing.T) {
	var teststruct teststruct
	err := New("testfiles/struct.yaml").ToStruct(&teststruct)
	assert.NoError(t, err)

	assert.NotNil(t, teststruct)
	assert.Equal(t, "test", teststruct.Test)
}

func TestHelper_YamlToStruct_nonexistent(t *testing.T) {
	var teststruct teststruct
	err := New("testfiles/nonexistent.yaml").ToStruct(&teststruct)
	assert.Error(t, err)
}
