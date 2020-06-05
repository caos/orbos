package helper

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelper_GetAbsPath(t *testing.T) {
	path, err := GetAbsPath("..", "..")
	assert.NoError(t, err)
	gopath := os.ExpandEnv("${GOPATH}")
	rootPath, err := GetAbsPath(gopath, "src", "github.com", "caos", "boom")
	assert.NoError(t, err)

	assert.Equal(t, path, rootPath)
}

func TestHelper_GetAbsPath_Nonexistent(t *testing.T) {
	path, err := GetAbsPath("nonexistent", "alsononexistent")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)

	path, err = GetAbsPath("nonexistent")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestHelper_RecreatePath(t *testing.T) {
	root := "/tmp/existent"
	err := os.MkdirAll(root, os.ModePerm)
	assert.NoError(t, err)
	assert.DirExists(t, root)

	err = ioutil.WriteFile("/tmp/existent/test.txt", []byte("test"), 0644)
	assert.NoError(t, err)
	files := getFiles(root)
	assert.Len(t, files, 1)

	err = RecreatePath(root)
	assert.NoError(t, err)
	assert.DirExists(t, root)

	files = getFiles(root)
	assert.Len(t, files, 0)

	err = os.RemoveAll(root)
	assert.NoError(t, err)
}

func TestHelper_RecreatePath_nonexistent(t *testing.T) {
	root := "/tmp/nonexistent"

	err := RecreatePath(root)
	assert.NoError(t, err)
	assert.DirExists(t, root)

	files := getFiles(root)
	assert.Len(t, files, 0)

	err = os.RemoveAll(root)
	assert.NoError(t, err)
}

func getFiles(root string) []string {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}
