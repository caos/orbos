package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func ensureORBITERTest(timeout time.Duration) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(_ newOrbctlCommandFunc, kubectl newKubectlCommandFunc) error {
		return watchLogs(kubectl, time.NewTimer(timeout))
	}
}

func watchLogs(kubectl newKubectlCommandFunc, timer *time.Timer) error {
	cmd := kubectl()
	cmd.Args = append(cmd.Args, "--namespace", "caos-system", "logs", "--follow", "--selector", "app.kubernetes.io/name=orbiter", "--since", "0s")
	cmd.Stderr = os.Stderr

	var success bool
	ensuredLog := "Desired state is ensured"

	err := simpleRunCommand(cmd, timer, func(line string) (goon bool) {
		fmt.Println(line)
		success = strings.Contains(line, ensuredLog)
		return !success
	})
	if success {
		return nil
	}
	if err != nil && !errors.Is(err, errTimeout) && !success {

		if err := checkORBITERRunning(kubectl); err != nil {
			return err
		}

		return watchLogs(kubectl, timer)
	}
	return err
}

func checkORBITERRunning(kubectl newKubectlCommandFunc) error {

	buf := new(bytes.Buffer)
	defer buf.Reset()

	cmd := kubectl()
	cmd.Args = append(cmd.Args, "--namespace", "caos-system", "get", "pods", "--selector", "app.kubernetes.io/name=orbiter", "--output", "yaml")
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("getting orbiter container status failed: %w", err)
	}

	data, err := ioutil.ReadAll(buf)
	if err != nil {
		return fmt.Errorf("reading orbiter container status failed: %w", err)
	}

	pods := struct {
		Items []struct {
			Status struct {
				ContainerStatuses []struct {
					State map[string]struct{}
				}
			}
		}
	}{}

	if err := yaml.Unmarshal(data, &pods); err != nil {
		return fmt.Errorf("unmarshalling orbiter container status failed: %w", err)
	}

	for i := range pods.Items {
		pod := pods.Items[i]
		for j := range pod.Status.ContainerStatuses {
			status := pod.Status.ContainerStatuses[j]
			if _, ok := status.State["waiting"]; ok {
				return nil
			}
			if _, ok := status.State["running"]; ok {
				return nil
			}
		}
	}

	return fmt.Errorf("ORBITER is not running")
}
