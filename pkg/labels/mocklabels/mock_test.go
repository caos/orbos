package mocklabels_test

import (
	"testing"

	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/labels/mocklabels"
)

func TestLabelMockTypes(t *testing.T) {

	for _, testcase := range []struct {
		labelType string
		typed     labels.Labels
		untyped   map[string]string
	}{{
		"Name",
		mocklabels.Name,
		mocklabels.NameMap,
	}, {
		"ClosedNameSelector",
		mocklabels.ClosedNameSelector,
		mocklabels.ClosedNameSelectorMap,
	}, {
		"Selectable",
		mocklabels.Selectable,
		mocklabels.SelectableMap,
	}} {
		typedStruct := labels.MustK8sMap(testcase.typed)

		for k, v := range typedStruct {
			untypedVal, ok := testcase.untyped[k]
			if !ok {
				t.Errorf("%s: key %s is missing in untyped labels", testcase.labelType, k)
				continue
			}

			if untypedVal != v {
				t.Errorf("%s: at key %s, typed value is %s, not %s", testcase.labelType, k, v, untypedVal)
			}
		}

		for k := range testcase.untyped {
			if _, ok := typedStruct[k]; !ok {
				t.Errorf("%s: key %s is expandable in untyped labels", testcase.labelType, k)
			}
		}
	}
}
