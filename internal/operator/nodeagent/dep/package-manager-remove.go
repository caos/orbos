package dep

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func (p *PackageManager) rembasedRemove(remove ...*Software) error {

	if len(remove) == 0 {
		return nil
	}

	swStrs := make([]string, len(remove))
	for i, sw := range remove {
		swStrs[i] = sw.Package
		if sw.Version != "" {
			swStrs[i] = fmt.Sprintf("%s-%s", sw.Package, sw.Version)
		}
	}

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	outBuf := new(bytes.Buffer)
	defer outBuf.Reset()

	cmd := exec.Command("yum", append([]string{"--assumeyes", "remove"}, swStrs...)...)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf
	err := cmd.Run()
	errStr := errBuf.String()
	outStr := outBuf.String()
	p.monitor.WithFields(map[string]interface{}{
		"command": fmt.Sprintf("'%s'", strings.Join(cmd.Args, "' '")),
		"stdout":  outStr,
		"stderr":  errStr,
	}).Debug("Executed yum remove")
	if err != nil {
		return fmt.Errorf("removing yum packages [%s] failed with stderr %s: %w", strings.Join(swStrs, ", "), errStr, err)
	}
	return nil
}
