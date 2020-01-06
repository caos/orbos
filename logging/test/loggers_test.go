// +build test unit

package test

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"github.com/caos/orbiter/logging/context"
	"github.com/caos/orbiter/logging/stdlib"
)

type testcase struct {
	desc string
	run  func() error
}

func TestLoggers(t *testing.T) {
	var buf bytes.Buffer
	context.Add(stdlib.New(&buf)).Verbose().WithFields(map[string]interface{}{
		"afield": "testfield",
		"amap": map[string]interface{}{
			"inner": "innertestfield",
		},
	}).Info("atestmessage")
	out := buf.Bytes()

	for _, expectedKeyPattern := range []string{
		"ts",
		"msg",
		"file",
		"line",
		"afield",
		"amap\\.inner",
	} {
		pattern := fmt.Sprintf(".*%s=.*\".*\".*", expectedKeyPattern)
		matches, err := regexp.Match(pattern, out)
		if err != nil {
			t.Error(err)
		}
		if !matches {
			t.Errorf("Expected %s to match pattern %s", string(out), pattern)
		}
	}
}
