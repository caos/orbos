package dynamic

import (
	"errors"
	"testing"

	"github.com/caos/orbos/v5/pkg/tree"
	"gopkg.in/yaml.v3"
)

func TestMigration(t *testing.T) {

	noWlTransport := map[string]interface{}{
		"name":       "ip1",
		"sourceport": 3000,
		"destinations": []map[string]interface{}{{
			"healthchecks": map[string]interface{}{
				"protocol": "http",
				"path":     "/",
				"code":     200,
			},
			"port": 33000,
			"pool": "dummy",
		}},
	}

	tests := []struct {
		name             string
		version          string
		spec             interface{}
		check            func(*Desired) error
		wantUnmarshalErr bool
	}{{
		name:    "Whitelists should be migrated from v0 to v1",
		version: "v0",
		spec: map[string][]map[string]interface{}{
			"pool": {{
				"ip": "192.168.0.10",
				"whitelist": []string{
					"127.168.0.20/32",
					"127.168.0.21/32",
				},
				"transport": []map[string]interface{}{noWlTransport, noWlTransport},
			}, {
				"ip": "192.168.0.11",
				"whitelist": []string{
					"127.168.0.20/32",
					"127.168.0.21/32",
					"127.168.0.22/32",
				},
				"transport": []map[string]interface{}{noWlTransport, noWlTransport},
			}},
		},
		check: func(desired *Desired) error {
			if desired.Common.Version != "v1" {
				return errors.New("Version not incremented")
			}
			if len(desired.Spec["pool"][0].Transport[0].Whitelist) != 2 {
				return errors.New("Whitelist not correctly moved")
			}
			if len(desired.Spec["pool"][0].Transport[1].Whitelist) != 2 {
				return errors.New("Whitelist not correctly moved")
			}
			if len(desired.Spec["pool"][1].Transport[0].Whitelist) != 3 {
				return errors.New("Whitelist not correctly moved")
			}
			return nil
		},
		wantUnmarshalErr: false,
	}, {
		name:             "V2 should not be supported",
		version:          "v2",
		spec:             map[string][]*interface{}{},
		check:            func(desired *Desired) error { return nil },
		wantUnmarshalErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := yaml.Marshal(map[string]interface{}{"spec": tt.spec})
			if err != nil {
				t.Fatal(err)
			}

			d := &Desired{Common: &tree.Common{Version: tt.version}}
			if unmarshalErr := yaml.Unmarshal(template, d); (unmarshalErr != nil) != tt.wantUnmarshalErr {
				t.Errorf("%s\nyaml.Unmarshal() error = %v, wantUnmarshalErr %v", string(template), unmarshalErr, tt.wantUnmarshalErr)
			}
			if err := tt.check(d); err != nil {
				t.Errorf("%s\nerror = %v", string(template), err)
			}
		})
	}
}
