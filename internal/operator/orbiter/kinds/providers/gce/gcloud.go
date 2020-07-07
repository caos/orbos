package gce

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var sdkDirCache string

var gcloudBinCache string

func gcloudSession(jsonkey string, do func(binary string) error) error {

	gcloud, err := exec.LookPath("gcloud")
	if err != nil {
		return err
	}

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

	cmd = exec.Command(gcloud, "auth", "activate-service-account", "--key-file", file.Name())
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := do(gcloud); err != nil {
		return err
	}
	if reactivate != "" {
		cmd := exec.Command(gcloud, "config", "configurations", "activate", reactivate)
		return cmd.Run()
	}
	return nil
}
