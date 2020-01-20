package common

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

func MarshalYAML(sth interface{}) []byte {
	if sth == nil {
		return nil
	}
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(sth); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
