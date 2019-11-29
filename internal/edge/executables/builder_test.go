package executables_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/caos/infrop/internal/edge/executables"
)

func TestBuildAndParseNodeAgent(t *testing.T) {

	/*	if err := os.Chdir("../../../../"); err != nil {
			t.Fatal(err)
		}
	*/
	testCommit := "itworks"
	testTag := "v0.1.0"

	cleanup, err := executables.Build("", "", testCommit, testTag)
	defer cleanup()

	if err != nil {
		t.Fatal(err)
	}

	packed, err := executables.Pack()
	if err != nil {
		t.Fatal(err)
	}

	if len(*packed) < 1 {
		t.Fatalf("packed nodeagent is empty")
	}

	unpacked, err := nodeagent.Unpack(packed)
	if err != nil {
		t.Fatal(err)
	}

	unpackedPath := filepath.Join(nodeagent.SelfPath, "unpacked")
	unpackedExecutable, err := os.Create(unpackedPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		unpackedExecutable.Close()
		os.Remove(unpackedPath)
	}()
	if err := unpackedExecutable.Chmod(0700); err != nil {
		t.Fatal(err)
	}

	if _, err = io.Copy(unpackedExecutable, bytes.NewReader(unpacked)); err != nil {
		t.Fatal(err)
	}
	unpackedExecutable.Close()

	var buf bytes.Buffer
	cmd := exec.Command(unpackedPath, "-version")
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	expected := fmt.Sprintf("%s %s\n", testTag, testCommit)
	actual := buf.String()
	if actual != expected {
		t.Fatalf("expected %s but got %s", expected, actual)
	}
}
