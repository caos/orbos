package core

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caos/orbos/internal/operator/common"

	"github.com/caos/orbos/internal/executables"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func NodeAgentFuncs(
	monitor mntr.Monitor,
	orbiterCommit string,
	repoURL string,
	repoKey string,
	currentNodeAgents *common.CurrentNodeAgents) (queryNodeAgent func(machine infra.Machine) (bool, error), install func(machine infra.Machine) error) {

	return func(machine infra.Machine) (running bool, err error) {

			machineMonitor := monitor.WithField("machine", machine.ID())

			var response []byte
			isActive := "sudo systemctl is-active node-agentd"
			err = infra.Try(monitor, time.NewTimer(7*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
				var cbErr error
				response, cbErr = cmp.Execute(nil, nil, isActive)
				return errors.Wrapf(cbErr, "remote command %s returned an unsuccessful exit code", isActive)
			})
			machineMonitor.WithFields(map[string]interface{}{
				"command":  isActive,
				"response": string(response),
			}).Debug("Executed command")
			if err != nil && !strings.Contains(string(response), "activating") {
				return false, nil
			}

			current, _ := currentNodeAgents.Get(machine.ID())
			if current.Commit == orbiterCommit {
				return true, nil
			}

			showVersion := "node-agent --version"
			err = infra.Try(monitor, time.NewTimer(7*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
				var cbErr error
				response, cbErr = cmp.Execute(nil, nil, showVersion)
				return errors.Wrapf(cbErr, "running command %s remotely failed", showVersion)
			})
			if err != nil {
				return false, err
			}
			machineMonitor.WithFields(map[string]interface{}{
				"command":  showVersion,
				"response": string(response),
			}).Debug("Executed command")

			fields := strings.Fields(string(response))
			return len(fields) > 1 && fields[1] == orbiterCommit, nil
		}, func(machine infra.Machine) error {

			machineMonitor := monitor.WithField("machine", machine.ID())

			systemdEntry := "node-agentd"
			systemdPath := fmt.Sprintf("/lib/systemd/system/%s.service", systemdEntry)
			nodeAgentPath := "/usr/local/bin/node-agent"
			healthPath := "/usr/local/bin/health"
			binary := nodeAgentPath
			if os.Getenv("MODE") == "DEBUG" {
				// Run node agent in debug mode
				if _, err := machine.Execute(nil, nil, "sudo apt-get update && sudo apt-get install -y git && wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz && sudo tar -zxvf go1.13.3.linux-amd64.tar.gz -C / && sudo chown -R $(id -u):$(id -g) /go && /go/bin/go get -u github.com/go-delve/delve/cmd/dlv && /go/bin/go install github.com/go-delve/delve/cmd/dlv && mv ${HOME}/go/bin/dlv /usr/local/bin"); err != nil {
					panic(err)
				}

				binary = fmt.Sprintf("dlv exec %s --api-version 2 --headless --listen 0.0.0.0:5000 --continue --accept-multiclient --", nodeAgentPath)
			}

			stopSystemd := fmt.Sprintf("sudo systemctl stop %s orbos.health* || true", systemdEntry)
			if err := infra.Try(monitor, time.NewTimer(8*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
				_, cbErr := cmp.Execute(nil, nil, stopSystemd)
				return errors.Wrapf(cbErr, "running command %s remotely failed", stopSystemd)
			}); err != nil {
				return errors.Wrap(err, "remotely stopping systemd services failed")
			}
			machineMonitor.WithField("command", stopSystemd).Debug("Executed command")

			writeFiles := []func() error{
				func() error {
					if err := infra.Try(monitor, time.NewTimer(8*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
						return errors.Wrapf(cmp.WriteFile(systemdPath, strings.NewReader(fmt.Sprintf(`[Unit]
Description=Node Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=%s --repourl "%s" --id "%s"
Restart=always
MemoryLimit=250M
RestartSec=10

[Install]
WantedBy=multi-user.target
`, binary, repoURL, machine.ID())), 600), "creating remote file %s failed", systemdPath)
					}); err != nil {
						return errors.Wrap(err, "remotely configuring Node Agent systemd unit failed")
					}
					machineMonitor.WithField("path", systemdPath).Debug("Written file")
					return nil
				},

				func() error {
					keyPath := "/etc/nodeagent/repokey"
					if err := infra.Try(monitor, time.NewTimer(8*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
						return errors.Wrapf(cmp.WriteFile(keyPath, strings.NewReader(repoKey), 400), "creating remote file %s failed", keyPath)
					}); err != nil {
						return errors.Wrap(err, "writing repokey failed")
					}
					machineMonitor.WithField("path", keyPath).Debug("Written file")
					return nil
				},

				func() error {
					nodeagent, err := executables.PreBuilt("nodeagent")
					if err != nil {
						return err
					}
					if err := infra.Try(monitor, time.NewTimer(20*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
						return errors.Wrapf(cmp.WriteFile(nodeAgentPath, bytes.NewReader(nodeagent), 700), "creating remote file %s failed", nodeAgentPath)
					}); err != nil {
						return errors.Wrap(err, "remotely installing Node Agent failed")
					}
					machineMonitor.WithField("path", nodeAgentPath).Debug("Written file")
					return nil
				},

				func() error {
					health, err := executables.PreBuilt("health")
					if err != nil {
						return err
					}
					if err := infra.Try(monitor, time.NewTimer(20*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
						return errors.Wrapf(cmp.WriteFile(healthPath, bytes.NewReader(health), 711), "creating remote file %s failed", healthPath)
					}); err != nil {
						return errors.Wrap(err, "remotely installing health executable failed")
					}
					machineMonitor.WithField("path", healthPath).Debug("Written file")
					return nil
				},
			}

			for _, fn := range writeFiles {
				if err := fn(); err != nil {
					return err
				}
			}

			restart := fmt.Sprintf("sudo systemctl daemon-reload && sudo systemctl enable %s && sudo systemctl restart %s", systemdPath, systemdEntry)
			if err := infra.Try(monitor, time.NewTimer(8*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
				_, cbErr := cmp.Execute(nil, nil, restart)
				return errors.Wrapf(cbErr, "running command %s remotely failed", restart)
			}); err != nil {
				return errors.Wrap(err, "reloading remote systemd failed")
			}
			machineMonitor.WithField("command", restart).Debug("Executed command")
			machineMonitor.Info("Node Agent installed")
			return nil
		}
}
