package labels_test

import (
	"strings"
	"testing"

	"github.com/caos/orbos/pkg/labels"
	"gopkg.in/yaml.v3"
)

func expectValidNameLabels(t *testing.T, name string) *labels.Name {
	l, err := labels.ForName(validComponentLabels(t), name)
	if err != nil {
		t.Fatal()
	}
	return l
}

func validNameLabels(t *testing.T) *labels.Name {
	return expectValidNameLabels(t, "testcase")
}

func TestNameLabels_Equal(t *testing.T) {
	expectValueEquality(
		t,
		validNameLabels(t),
		validNameLabels(t),
		expectValidNameLabels(t, "somethingElse"))
}

const validLabels = `app.kubernetes.io/name: testcase
app.kubernetes.io/component: testSuite
caos.ch/kind: testing.caos.ch/TestSuite
caos.ch/apiversion: v1
app.kubernetes.io/managed-by: TEST_OPERATOR_LABELS
app.kubernetes.io/version: v123.4.5
app.kubernetes.io/part-of: ORBOS
`

func TestNameLabels_MarshalYAML(t *testing.T) {
	marshalled, err := yaml.Marshal(validNameLabels(t))
	if err != nil {
		t.Error(err, "expected successful mashalling")
	}

	if strings.TrimSpace(string(marshalled)) != strings.TrimSpace(validLabels) {
		t.Errorf("expected \n%s\n but got \n%s\n", validLabels, string(marshalled))
	}
}

func TestNameLabels_UnmarshalYAML(t *testing.T) {
	name := &labels.Name{}
	if err := yaml.Unmarshal([]byte(validLabels), name); err != nil {
		t.Fatal(err)
	}

	reMarshalled, err := yaml.Marshal(name)
	if err != nil {
		t.Fatal(err)
	}

	got := string(reMarshalled)
	if got != validLabels {
		t.Errorf("expected %s but got %s", validLabels, got)
	}
}
