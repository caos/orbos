package gce

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func gcloudSession(jsonkey string, gcloud string, do func(binary string) error) error {

	listBuf := new(bytes.Buffer)
	defer resetBuffer(listBuf)
	cmd := exec.Command(gcloud, "config", "configurations", "list")
	cmd.Stdout = listBuf

	if err := cmd.Run(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(listBuf)
	reactivate := ""
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if fields[1] == "True" {
			reactivate = fields[0]
			break
		}
	}

	if err := run(exec.Command(gcloud, "config", "configurations", "create", "orbiter-system"), func(out []byte) bool {
		return strings.Contains(string(out), "it already exists")
	}); err != nil {
		return err
	}

	if err := run(exec.Command(gcloud, "config", "configurations", "activate", "orbiter-system"), nil); err != nil {
		return err
	}

	file, err := ioutil.TempFile("", "orbiter-gce-key")
	defer os.Remove(file.Name())
	if err != nil {
		return err
	}

	_, err = file.WriteString(jsonkey)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	if err := run(exec.Command(gcloud, "auth", "activate-service-account", "--key-file", file.Name()), nil); err != nil {
		return err
	}
	if err := do(gcloud); err != nil {
		return err
	}
	if reactivate != "" {
		if err := run(exec.Command(gcloud, "config", "configurations", "activate", reactivate), nil); err != nil {
			return err
		}
	}
	return nil
}

func run(cmd *exec.Cmd, ignoreErr func([]byte) bool) error {
	out, err := cmd.CombinedOutput()
	if err != nil && ignoreErr != nil && !ignoreErr(out) {
		return fmt.Errorf("failed to run \"%s\": %s: %w", strings.Join(cmd.Args, "\" \""), string(out), err)
	}
	return nil
}
