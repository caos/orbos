package yaml

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type yamlFile struct {
	path string
}

func New(path string) *yamlFile {
	return &yamlFile{
		path: path,
	}
}

func readFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while reading yaml %s", path)
	}

	return data, nil
}

func (y *yamlFile) DeleteFile() error {
	return os.Remove(y.path)
}

func (y *yamlFile) ToString() (string, error) {
	data, err := readFile(y.path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (y *yamlFile) ToStruct(struc interface{}) error {
	parts, err := y.ToObjectList()
	if err != nil {
		return err
	}

	for _, part := range parts {
		if part == "" {
			continue
		}
		err = yaml.Unmarshal([]byte(part), struc)
		if err != nil {
			return errors.Wrapf(err, "Error while unmarshaling yaml %s to struct", y.path)
		}
		return nil
	}
	return nil
}

func (y *yamlFile) ToObjectList() ([]string, error) {
	data, err := readFile(y.path)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), "\n---\n"), nil
}

func (y *yamlFile) AddStruct(struc interface{}) error {
	data, err := yaml.Marshal(struc)
	if err != nil {
		return err
	}

	if err := writeFileYAMLDivider(y.path); err != nil {
		return err
	}

	return writeFile(y.path, data)
}

func (y *yamlFile) AddString(str string) error {
	return writeFile(y.path, []byte(str))
}

func (y *yamlFile) AddStringObject(str string) error {
	if err := writeFileYAMLDivider(y.path); err != nil {
		return err
	}

	return writeFile(y.path, []byte(str))
}

func writeFileYAMLDivider(path string) error {
	if fileExists(path) {
		f, err := os.OpenFile(path,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		defer f.Close()

		if _, err := f.WriteString("\n---\n"); err != nil {
			return err
		}
	}

	return nil
}

func writeFile(path string, data []byte) error {
	f, err := os.OpenFile(path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
