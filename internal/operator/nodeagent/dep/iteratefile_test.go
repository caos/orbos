package dep_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/caos/orbos/internal/operator/nodeagent/dep"
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

	writer := new(bytes.Buffer)
	defer writer.Reset()

	if err := dep.Manipulate(reader, writer, []string{"second"}, []string{"fourth line"}, func(line string) *string {
		if strings.Contains(line, "third") {
			str := "#" + line
			return &str
		}
		return &line
	}); err != nil {
		t.Fatal(err)
	}

	result := writer.String()
	if result != expect {
		t.Errorf("Expected:\n%s\n\nGenerated:\n%s", expect, result)
	}
}
