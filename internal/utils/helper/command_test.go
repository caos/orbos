package helper

import (
	"os/exec"
	"testing"

	"github.com/caos/orbos/mntr"
	"github.com/stretchr/testify/assert"
)

func newMonitor() mntr.Monitor {

	monitor := mntr.Monitor{
		OnInfo:   mntr.LogMessage,
		OnChange: mntr.LogMessage,
		OnError:  mntr.LogError,
	}
	return monitor
}

func TestHelper_Run(t *testing.T) {
	monitor := newMonitor()

	cmd := exec.Command("echo", "first")
	err := Run(monitor, *cmd)
	assert.NoError(t, err)
}

func TestHelper_Run_MoreArgs(t *testing.T) {
	monitor := newMonitor()

	cmd := exec.Command("echo", "first", "second")
	err := Run(monitor, *cmd)
	assert.NoError(t, err)
}

func TestHelper_Run_UnknowCommand(t *testing.T) {
	monitor := newMonitor()

	cmd := exec.Command("unknowncommand", "first")
	err := Run(monitor, *cmd)
	assert.Error(t, err)
}

func TestHelper_Run_ErrorCommand(t *testing.T) {
	monitor := newMonitor()

	cmd := exec.Command("ls", "/unknownfolder/unknownsubfolder")
	err := Run(monitor, *cmd)
	assert.Error(t, err)
}
