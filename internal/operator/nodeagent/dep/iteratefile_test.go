package dep_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/caos/orbiter/internal/operator/nodeagent/dep"
)

func TestIterateFile(t *testing.T) {

	from := `first line
second line
third line
`

	expect := `first line
#third line
fourth line
`

	reader := strings.NewReader(from)

	var writer bytes.Buffer

	if err := dep.Manipulate(reader, &writer, []string{"second"}, []string{"fourth line"}, func(line string) string {
		if strings.Contains(line, "third") {
			return "#" + line
		}
		return line
	}); err != nil {
		t.Fatal(err)
	}

	result := writer.String()
	if result != expect {
		t.Errorf("Expected:\n%s\n\nGenerated:\n%s", expect, result)
	}
}
