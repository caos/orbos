package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/afiskon/promtail-client/promtail"

	"gopkg.in/yaml.v3"
)

func ensureORBITERTest(logger promtail.Client, timeout time.Duration, condition func(promtail.Client, newOrbctlCommandFunc) (bool, error)) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, kubectl newKubectlCommandFunc) error {
		return watchLogs(logger, orbctl, kubectl, time.NewTimer(timeout), condition)
	}
}

func watchLogs(logger promtail.Client, orbctl newOrbctlCommandFunc, kubectl newKubectlCommandFunc, timer *time.Timer, condition func(promtail.Client, newOrbctlCommandFunc) (bool, error)) error {
	cmd := kubectl()
	cmd.Args = append(cmd.Args, "--namespace", "caos-system", "logs", "--follow", "--selector", "app.kubernetes.io/name=orbiter", "--since", "10s")

	errWriter, errWrite := logWriter(logger.Warnf)
	defer errWrite()
	cmd.Stderr = errWriter

	var success bool
	ensuredLog := "Desired state is ensured"

	err := simpleRunCommand(cmd, timer, func(line string) (goon bool) {
		log := logger.Infof
		if strings.Contains(line, ` err="`) {
			log = logger.Warnf
		}
		log(line)
		success = strings.Contains(line, ensuredLog)
		return !success
	})
	if success {
		ok, err := condition(logger, orbctl)
		if err != nil {
			return err
		}
		if !ok {
			return watchLogs(logger, orbctl, kubectl, timer, condition)
		}
		return nil
	}

	if errors.Is(err, errTimeout) {
		return err
	}
	err = nil

	// give orbiter two minutes to become waiting or running
	minute := time.NewTimer(2 * time.Minute)
	for {
		select {
		case <-minute.C:
			return errors.New("orbiter wasn't running for a minute")
		default:
			if err := checkORBITERRunning(logger, kubectl); err != nil {
				logger.Warnf(err.Error())
				continue
			}
			break
		}
	}
	return watchLogs(logger, orbctl, kubectl, timer, condition)
}

func checkORBITERRunning(logger promtail.Client, kubectl newKubectlCommandFunc) error {

	buf := new(bytes.Buffer)
	defer buf.Reset()

	cmd := kubectl()
	cmd.Args = append(cmd.Args, "--namespace", "caos-system", "get", "pods", "--selector", "app.kubernetes.io/name=orbiter", "--output", "yaml")
	cmd.Stdout = buf

	errWriter, errWrite := logWriter(logger.Warnf)
	defer errWrite()
	cmd.Stderr = errWriter

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
