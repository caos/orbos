// +build test unit

package core_test

import (
	"errors"
	"os"
	"testing"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/logging/base"
	logcontext "github.com/caos/orbiter/logging/context"
)

type equalsCase struct {
	desc   string
	first  *core.Resource
	second *core.Resource
	equals bool
}

type mockService struct {
	IrrelevantConfig interface{}
}

func (m *mockService) Abbreviate() string {
	return "mock"
}

func (m *mockService) Desire(payload interface{}) (interface{}, error) {
	return payload, nil
}

func (m *mockService) Ensure(id string, desired interface{}, dependencies []interface{}) (interface{}, error) {
	return nil, errors.New("Not ensureable")
}

func (m *mockService) Delete(id string) error {
	return errors.New("Not deletable")
}

func (m *mockService) AllExisting() ([]string, error) {
	return nil, errors.New("Not queriable")
}

func createInner(cfgValue string) struct{ Anything struct{ Inner string } } {
	return struct{ Anything struct{ Inner string } }{struct{ Inner string }{cfgValue}}
}

func createResource(svc core.ResourceService, payloadValue *string, deps []*core.Resource) *core.Resource {
	var cfg interface{}
	if payloadValue != nil {
		cfg = createInner(*payloadValue)
	}
	return core.NewResourceFactory(logcontext.Add(stdlib.New(os.Stdout)), "testoperator").New(svc, cfg, deps, nil)
}

func TestResourcesEquality(t *testing.T) {

	first := "first"
	second := "second"

	for idx, args := range []*equalsCase{
		&equalsCase{
			desc:   "Resources with same configuration should equal",
			first:  createResource(&mockService{}, &first, nil),
			second: createResource(&mockService{}, &first, nil),
			equals: true,
		},
		&equalsCase{
			desc:   "Resources with different configuration should not equal",
			first:  createResource(&mockService{}, &first, nil),
			second: createResource(&mockService{}, &second, nil),
			equals: false,
		},
		&equalsCase{
			desc: "Resources with same configuration but different dependencies should not equal",
			first: createResource(&mockService{}, &first, []*core.Resource{
				createResource(&mockService{}, &first, nil),
				createResource(&mockService{}, &second, nil),
			}),
			second: createResource(&mockService{}, &first, []*core.Resource{
				createResource(&mockService{}, &first, nil),
			}),
			equals: false,
		},
		&equalsCase{
			desc: "Dependencies order should not matter",
			first: createResource(&mockService{}, &first, []*core.Resource{
				createResource(&mockService{}, &first, nil),
				createResource(&mockService{}, &second, nil),
			}),
			second: createResource(&mockService{}, &first, []*core.Resource{
				createResource(&mockService{}, &second, nil),
				createResource(&mockService{}, &first, nil),
			}),
			equals: true,
		},

		// Only Resources should be compared, not ResourceServices as
		// services possibly reconfigure themselves - e.g. for caching
		&equalsCase{
			desc:   "Dependencies order should not matter",
			first:  createResource(&mockService{createInner("first")}, &first, nil),
			second: createResource(&mockService{createInner("second")}, &first, nil),
			equals: true,
		},
	} {

		firstID, err := args.first.ID()
		if err != nil {
			t.Error(err)
		}
		secondID, err := args.second.ID()
		if err != nil {
			t.Error(err)
		}

		if (*firstID == *secondID) != args.equals {
			expect := ""
			if !args.equals {
				expect = " not"
			}
			t.Logf("%d) %s\n", idx+1, args.desc)
			t.Logf("%d) Expecting %s %+v to%s equal %s %+v\n", idx+1, *firstID, args.first, expect, *secondID, args.second)
			t.Error()
		}
	}
}
