package tree

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCommon_MarshalYAML(t *testing.T) {

	tests := []struct {
		name             string
		parsedApiVersion bool
		wantKey          string
	}{{
		name:             "It should marshal kind and normal version",
		parsedApiVersion: false,
		wantKey:          "version",
	}, {
		name:             "It should marshal kind and api version using the version field",
		parsedApiVersion: true,
		wantKey:          "apiVersion",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := yaml.Marshal(NewCommon("akind", "aversion", tt.parsedApiVersion))
			if err != nil {
				t.Error(err)
				return
			}
			gotStr := string(got)
			wantStr := fmt.Sprintf(`kind: akind
%s: aversion
`, tt.wantKey)

			if gotStr != wantStr {
				t.Errorf("MarshalYAML() got = %v, want %v", gotStr, wantStr)
			}
		})
	}
}
