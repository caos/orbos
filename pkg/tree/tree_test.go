package tree

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCommon_MarshalYAML(t *testing.T) {
	type fields struct {
		Kind             string
		Version          string
		X_ApiVersion     string
		parsedApiVersion bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{{
		name: "It should marshal kind and normal version",
		fields: fields{
			Kind:             "akind",
			Version:          "aversion",
			X_ApiVersion:     "",
			parsedApiVersion: false,
		},
		want: `kind: akind
version: aversion
`,
		wantErr: false,
	}, {
		name: "It should marshal kind and api version using the version field",
		fields: fields{
			Kind:             "akind",
			Version:          "aversion",
			X_ApiVersion:     "ignore",
			parsedApiVersion: true,
		},
		want: `kind: akind
apiVersion: aversion
`,
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Common{
				Kind:             tt.fields.Kind,
				Version:          tt.fields.Version,
				X_ApiVersion:     tt.fields.X_ApiVersion,
				parsedApiVersion: tt.fields.parsedApiVersion,
			}

			got, err := yaml.Marshal(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotStr := string(got)
			if gotStr != tt.want {
				t.Errorf("MarshalYAML() got = %v, want %v", gotStr, tt.want)
			}
		})
	}
}
