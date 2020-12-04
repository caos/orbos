package labels_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbos/pkg/labels"
)

func expectValidOperatorLabels(t *testing.T, operator, version string) *labels.Operator {
	l, err := labels.ForOperator("ORBOS", operator, version)
	if err != nil {
		t.Fatal()
	}
	return l
}

func validOperatorLabels(t *testing.T) *labels.Operator {
	return expectValidOperatorLabels(t, "TEST_OPERATOR_LABELS", "testing-dev")
}

func TestOperatorLabels_Equal(t *testing.T) {
	expectValueEquality(
		t,
		validOperatorLabels(t),
		validOperatorLabels(t),
		expectValidOperatorLabels(t, "TWO", "testing-dev"),
	)
}

func TestOperatorLabels_MarshalYAML(t *testing.T) {
	testCase := validOperatorLabels(t)
	_, err := yaml.Marshal(testCase)
	if err == nil {
		t.Error("expected full set of labels")
	}
}

func TestOperatorLabels_Major(t *testing.T) {
	expectUnknown := "non-semver version should return unknown major"
	expectCorrect := "semver version should return correct major"
	testcases := []struct {
		msg     string
		version string
		major   int8
	}{{
		msg:     expectUnknown,
		version: " v1.2.3",
		major:   -1,
	}, {
		msg:     expectUnknown,
		version: "v1.2.3 ",
		major:   -1,
	}, {
		msg:     expectUnknown,
		version: "v1.2-3",
		major:   -1,
	}, {
		msg:     expectUnknown,
		version: "1.2.3",
		major:   -1,
	}, {
		msg:     expectUnknown,
		version: "v9999.2.3",
		major:   -1,
	}, {
		msg:     expectCorrect,
		version: "v1.2.3",
		major:   1,
	}, {
		msg:     expectCorrect,
		version: "v0.2.3",
		major:   0,
	}, {
		msg:     expectCorrect,
		version: "v123.2.3",
		major:   123,
	}}

	for idx := range testcases {
		tCase := testcases[idx]
		major := expectValidOperatorLabels(t, "ORBOS", tCase.version).Major()
		if major != tCase.major {
			t.Errorf("%s: expected major %d for label %s but got %d", tCase.msg, tCase.major, tCase.version, major)
		}
	}
}
