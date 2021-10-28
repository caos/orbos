package host

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func Test_contains(t *testing.T) {

	unmarshal := func(from string) map[string]interface{} {
		to := make(map[string]interface{})
		if err := yaml.Unmarshal([]byte(from), to); err != nil {
			t.Fatalf("unmarshalling failed: %v: %s", err, from)
		}
		return to
	}

	tests := []struct {
		name   string
		subset string
		want   bool
	}{{
		name: "it should contain",
		subset: `
apiVersion: getambassador.io/v2
kind: Host
metadata:
  annotations:
    aes_res_changed: "true"
  labels:
    caos.ch/apiversion: v0
    caos.ch/kind: ZITADEL
  name: accounts
  namespace: caos-zitadel
spec:
  acmeProvider:
    authority: https://acme-v02.api.letsencrypt.org/directory
  ambassador_id:
  - default
  hostname: accounts.iam.sustema.ai
  selector:
    matchLabels:
      hostname: accounts.iam.sustema.ai
`,
		want: true,
	}, {
		name: "it should not contain",
		subset: `
apiVersion: getambassador.io/v2
kind: Host
metadata:
  annotations:
    aes_res_changed: "true"
  labels:
    caos.ch/apiversion: v0
    caos.ch/kind: ZITADEL
  name: accounts
  namespace: caos-zitadel
spec:
  acmeProvider:
    authority: https://acme-v02.api.letsencrypt.org/directory
  ambassador_id:
  - default
  - blubb ####################################################################
  hostname: accounts.iam.sustema.ai
  selector:
    matchLabels:
      hostname: accounts.iam.sustema.ai
  tlsSecret:
    name: accounts.iam.sustema.ai
`,
		want: false,
	}}

	existing := unmarshal(`
apiVersion: getambassador.io/v2
kind: Host
metadata:
  annotations:
    aes_res_changed: "true"
  creationTimestamp: "2021-04-21T15:04:55Z"
  generation: 9
  labels:
    caos.ch/apiversion: v0
    caos.ch/kind: ZITADEL
  managedFields:
  - apiVersion: getambassador.io/v2
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .: {}
          f:aes_res_changed: {}
        f:labels:
          .: {}
          f:app.kubernetes.io/component: {}
    manager: zitadelctl
  name: accounts
  namespace: caos-zitadel
  resourceVersion: "487823"
  selfLink: /apis/getambassador.io/v2/namespaces/caos-zitadel/hosts/accounts
  uid: ffa02ea0-6b40-43d9-abd8-53e135e98e6a
spec:
  acmeProvider:
    authority: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecret:
      name: https-3a-2f-2facme-2dv02.api.letsencrypt.org-2fdirectory
    registration: '{"body":{"status":"valid"},"uri":"https://acme-v02.api.letsencrypt.org/acme/acct/120159215"}'
  ambassador_id:
  - default
  hostname: accounts.iam.sustema.ai
  selector:
    matchLabels:
      hostname: accounts.iam.sustema.ai
  tlsSecret:
    name: accounts.iam.sustema.ai
status:
  state: Ready
  tlsCertificateSource: ACME
`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := contains(existing, unmarshal(tt.subset)); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
