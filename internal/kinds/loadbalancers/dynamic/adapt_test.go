package dynamic_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic"
)

func TestDynamic(t *testing.T) {

	rawDesired := `kind: orbiter.caos.ch/DynamicLoadBalancer
version: v0
spec:
    masters:
      - ip: 10.73.245.69
        transport:
          - name: kubeapi
            sourcePort: 6443
            destinations:
              - pool: masters
                port: 6443
                healthchecks:
                    code: 200
                    path: /healthz
                    protocol: https
    workers:
      - ip: 10.73.245.70
        transport:
          - name: httpsingress
            sourcePort: 443
            destinations:
              - pool: workers
                port: 30443
                healthchecks:
                    code: 200
                    path: /ambassador/v0/check_ready
                    protocol: https
          - name: httpingress
            sourcePort: 80
            destinations:
              - pool: workers
                port: 30080
                healthchecks:
                    code: 200
                    path: /ambassador/v0/check_ready
                    protocol: http`

	treeDesired := &orbiter.Tree{}
	if err := yaml.Unmarshal([]byte(rawDesired), treeDesired); err != nil {
		t.Fatal(err)
	}

	treeCurrent := &orbiter.Tree{}
	if _, err := dynamic.AdaptFunc()(treeDesired, nil, treeCurrent); err != nil {
		t.Fatal(err)
	}

	marshaledDesired, err := yaml.Marshal(treeDesired)
	if err != nil {
		t.Fatal(err)
	}

	if string(marshaledDesired) != rawDesired {
		t.Errorf("\nactual desired...\n%s\n\nexpected desired...\n%s", string(marshaledDesired), rawDesired)
	}

}
