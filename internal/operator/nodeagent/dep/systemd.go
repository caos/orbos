package dep

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/caos/orbos/mntr"
)

type SystemD struct {
	monitor mntr.Monitor
}

func NewSystemD(monitor mntr.Monitor) *SystemD {
	return &SystemD{monitor}
}

func (s *SystemD) Disable(binary string) error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("systemctl", "stop", binary)
	cmd.Stderr = errBuf
	if s.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	err := cmd.Run()
	if err != nil {
		errString := errBuf.String()
		if strings.Contains(errString, "not loaded") {
			err = nil
		} else {
			return fmt.Errorf("stopping %s by systemd failed with stderr %s: %w", binary, errString, err)
		}
	}

	errBuf.Reset()
	cmd = exec.Command("systemctl", "disable", binary)
	cmd.Stderr = errBuf
	if s.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		errString := errBuf.String()
		if strings.Contains(errString, "No such file or directory") {
			err = nil
		} else {
			return fmt.Errorf("disabling %s by systemd failed with stderr %s: %w", binary, errString, err)
		}
	}

	return nil
}

func (s *SystemD) Start(binary string) error {
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("systemctl", "restart", binary)
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restarting %s from systemd failed with stderr %s: %w", binary, errBuf.String(), err)
	}
	return nil
}

func (s *SystemD) Enable(binary string) error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("systemctl", "enable", binary)
	cmd.Stderr = errBuf
	if s.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("enabling systemd unit %s failed with stderr %s", binary, errBuf.String(), err)
	}

	if !s.Active(binary) {
		return s.Start(binary)
	}
	return nil
}

func (s *SystemD) Active(binary string) bool {
	cmd := exec.Command("systemctl", "is-active", binary)
	if s.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	return cmd.Run() == nil
}
