package helper

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/caos/orbos/v5/internal/utils/yaml"

	"github.com/stretchr/testify/assert"
)

var (
	firstResource = &Resource{
		Kind:       "test1",
		ApiVersion: "api/v1",
		Metadata: &Metadata{
			Name:      "test1",
			Namespace: "test",
		},
	}
	secondResource = &Resource{
		Kind:       "test2",
		ApiVersion: "api/v1",
		Metadata: &Metadata{
			Name:      "test2",
			Namespace: "test",
		},
	}
)

type metadataTest struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}
type resourceTest struct {
	Kind       string        `yaml:"kind"`
	ApiVersion string        `yaml:"apiVersion"`
	Metadata   *metadataTest `yaml:"metadata"`
}

func TestHelper_AddStructToYaml(t *testing.T) {
	root := "/tmp/nonexistent"
	err := os.MkdirAll(root, os.ModePerm)
	assert.NoError(t, err)

	path := "/tmp/nonexistent/test.yaml"
	err = yaml.New(path).AddStruct(firstResource)
	assert.NoError(t, err)

	files := getFiles(root)
	assert.Len(t, files, 1)

	var restTest Resource
	err = yaml.New(path).ToStruct(&restTest)
	assert.NoError(t, err)

	assert.True(t, reflect.DeepEqual(firstResource, &restTest))

	err = os.RemoveAll(root)
	assert.NoError(t, err)
}

func TestHelper_AddStringToYaml(t *testing.T) {
	root := "/tmp/nonexistent"
	err := os.MkdirAll(root, os.ModePerm)
	assert.NoError(t, err)

	path := "/tmp/nonexistent/test.yaml"
	err = yaml.New(path).AddString("test")
	assert.NoError(t, err)

	files := getFiles(root)
	assert.Len(t, files, 1)

	restTest, err := yaml.New(path).ToString()
	assert.NoError(t, err)

	assert.Equal(t, "test", restTest)

	err = os.RemoveAll(root)
	assert.NoError(t, err)
}

func TestHelper_AddStringObjectToYaml(t *testing.T) {
	root := "/tmp/nonexistent"
	err := os.MkdirAll(root, os.ModePerm)
	assert.NoError(t, err)

	path := "/tmp/nonexistent/test.yaml"
	y := yaml.New(path)
	err = y.AddStringObject("test: test")
	assert.NoError(t, err)

	files := getFiles(root)
	assert.Len(t, files, 1)

	restTest, err := y.ToString()
	assert.NoError(t, err)
	assert.Equal(t, "test: test", restTest)

	err = os.RemoveAll(root)
	assert.NoError(t, err)
}

func TestHelper_AddStringBeforePointForKindAndName(t *testing.T) {
	root := "/tmp/nonexistent"
	err := os.MkdirAll(root, os.ModePerm)
	assert.NoError(t, err)

	y := yaml.New("/tmp/nonexistent/test.yaml")
	err = y.AddStruct(firstResource)
	assert.NoError(t, err)

	firstResourceTest := &resourceTest{
		Kind:       firstResource.Kind,
		ApiVersion: firstResource.ApiVersion,
		Metadata: &metadataTest{
			Name:      firstResource.Metadata.Name,
			Namespace: "test",
		},
	}

	files := getFiles(root)
	assert.Len(t, files, 1)

	var restTest resourceTest
	err = y.ToStruct(&restTest)
	assert.NoError(t, err)

	assert.True(t, reflect.DeepEqual(firstResourceTest, &restTest))

	err = os.RemoveAll(root)
	assert.NoError(t, err)
}

func prepare(t *testing.T, dir, filename string) string {
	path := filepath.Join(dir, filename)
	err := os.MkdirAll(dir, os.ModePerm)
	assert.NoError(t, err)

	y := yaml.New(path)
	err = y.AddStruct(firstResource)
	assert.NoError(t, err)
	err = y.AddStruct(secondResource)
	assert.NoError(t, err)

	files := getFiles(dir)
	assert.Len(t, files, 1)

	return path
}

func cleanup(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	assert.NoError(t, err)
}

func TestHelper_DeleteKindFromYaml_first(t *testing.T) {
	root := "/tmp/nonexistent"
	path := prepare(t, root, "test.yaml")

	err := DeleteKindFromYaml(path, "test1")
	assert.NoError(t, err)

	var restTest Resource
	err = yaml.New(path).ToStruct(&restTest)
	assert.NoError(t, err)

	assert.True(t, reflect.DeepEqual(secondResource, &restTest))

	cleanup(t, root)
}

func TestHelper_DeleteKindFromYaml_second(t *testing.T) {
	root := "/tmp/nonexistent"
	path := prepare(t, root, "test.yaml")

	err := DeleteKindFromYaml(path, "test2")
	assert.NoError(t, err)

	var restTest Resource
	err = yaml.New(path).ToStruct(&restTest)
	assert.NoError(t, err)

	assert.True(t, reflect.DeepEqual(firstResource, &restTest))

	cleanup(t, root)
}

func TestHelper_DeleteKindFromYaml_both(t *testing.T) {
	root := "/tmp/nonexistent"
	path := prepare(t, root, "test.yaml")

	err := DeleteKindFromYaml(path, "test1")
	assert.NoError(t, err)

	err = DeleteKindFromYaml(path, "test2")
	assert.NoError(t, err)

	var restTest Resource
	err = yaml.New(path).ToStruct(&restTest)
	assert.Error(t, err)

	assert.Empty(t, &restTest)

	cleanup(t, root)
}

func TestHelper_DeleteFirstResourceFromYaml_first(t *testing.T) {
	root := "/tmp/nonexistent"
	path := prepare(t, root, "test.yaml")

	err := DeleteFirstResourceFromYaml(path, "api/v1", "test1", "test1")
	assert.NoError(t, err)

	var restTest Resource
	err = yaml.New(path).ToStruct(&restTest)
	assert.NoError(t, err)

	assert.True(t, reflect.DeepEqual(secondResource, &restTest))

	cleanup(t, root)
}
