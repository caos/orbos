package core

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caos/orbos/internal/executables"
	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	orbcfg "github.com/caos/orbos/pkg/orb"
)

type IterateNodeAgentFuncs func(currentNodeAgents *common.CurrentNodeAgents) (queryNodeAgent func(machine infra.Machine, orbiterCommit string) (bool, error), install func(machine infra.Machine) error)

func ConfigureNodeAgents(svc MachinesService, monitor mntr.Monitor, orb orbcfg.Orb, pprof bool) error {
	configure, _ := NodeAgentFuncs(monitor, orb.URL, orb.Repokey, pprof)
	return Each(svc, func(pool string, machine infra.Machine) error {
		err := configure(machine)
		if err != nil {
			return fmt.Errorf("configuring node agent on machine %s in pool %s failed: %w", machine.ID(), pool, err)
		}
		monitor.WithFields(map[string]interface{}{
			"pool":     pool,
			"machine:": machine.ID(),
		}).Info("Node agent configured")
		return nil
	})
}

func NodeAgentFuncs(
	monitor mntr.Monitor,
	repoURL string,
	repoKey string,
	pprof bool,
) (
	reconfigure func(machines infra.Machine) error,
	iterate IterateNodeAgentFuncs,
) {

	configure := func(machine infra.Machine) func() error {
		machineMonitor := monitor.WithField("machine", machine.ID())
		return helpers.Fanout([]func() error{
			func() error {
				keyPath := "/var/orbiter/repo-key"
				if err := infra.Try(machineMonitor, time.NewTimer(8*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
					if cbErr := cmp.WriteFile(keyPath, strings.NewReader(repoKey), 400); cbErr != nil {
						return fmt.Errorf("creating remote file %s failed: %w", keyPath, cbErr)
					}
					return nil
				}); err != nil {
					return fmt.Errorf("writing repokey failed: %w", err)
				}
				machineMonitor.WithFields(map[string]interface{}{
					"path": keyPath,
				}).Debug("Written file")
				return nil
			},
			func() error {
				urlPath := "/var/orbiter/repo-url"
				if err := infra.Try(machineMonitor, time.NewTimer(8*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
					if cbErr := cmp.WriteFile(urlPath, strings.NewReader(repoURL), 400); cbErr != nil {
						return fmt.Errorf("creating remote file %s failed: %w", urlPath, cbErr)
					}
					return nil
				}); err != nil {
					return fmt.Errorf("writing repourl failed: %w", err)
				}
				machineMonitor.WithFields(map[string]interface{}{
					"path": urlPath,
				}).Debug("Written file")
				return nil
			},
		})
	}

	nodeAgentPath := "/usr/local/bin/node-agent"
	binary := nodeAgentPath
	pprofStr := ""
	if pprof {
		pprofStr = "--pprof"
	}
	verboseStr := ""
	if monitor.IsVerbose() {
		verboseStr = "--verbose"
	}

	sentryEnvironment, _, enabled := mntr.Environment()
	if !enabled {
		sentryEnvironment = ""
	}

	systemdEntry := "node-agentd"
	systemdPath := fmt.Sprintf("/lib/systemd/system/%s.service", systemdEntry)
	systemdUnitCache := make(map[string]string)
	systemdUnitFile := func(machine infra.Machine) string {

		if cached, ok := systemdUnitCache[machine.ID()]; ok {
			return cached
		}

		newFile := fmt.Sprintf(`[Unit]
Description=Node Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=%s --id "%s" %s %s --environment "%s"
Restart=on-failure
MemoryLimit=1G
MemoryAccounting=yes
RestartSec=10
CPUAccounting=yes
TimeoutStopSec=300
KillMode=mixed

[Install]
WantedBy=multi-user.target
`, binary, machine.ID(), pprofStr, verboseStr, sentryEnvironment)
		systemdUnitCache[machine.ID()] = newFile
		return newFile
	}

	return func(machine infra.Machine) error {
			return configure(machine)()
		}, func(currentNodeAgents *common.CurrentNodeAgents) (func(infra.Machine, string) (bool, error), func(infra.Machine) error) {

			return func(machine infra.Machine, orbiterCommit string) (running bool, err error) {

					machineMonitor := monitor.WithField("machine", machine.ID())

					var response []byte
					isActive := "sudo systemctl is-active node-agentd"
					err = infra.Try(machineMonitor, time.NewTimer(7*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
						var cbErr error
						if response, cbErr = cmp.Execute(nil, isActive); cbErr != nil {
							return fmt.Errorf("remote command %s returned an unsuccessful exit code: %w", isActive, cbErr)
						}
						return nil
					})
					machineMonitor.WithFields(map[string]interface{}{
						"command":  isActive,
						"response": string(response),
					}).Debug("Executed command")
					if err != nil && !strings.Contains(string(response), "activating") {
						return false, nil
					}

					remoteSystemdUnitFile := new(bytes.Buffer)
					defer remoteSystemdUnitFile.Reset()
					err = infra.Try(machineMonitor, time.NewTimer(7*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
						if err := cmp.ReadFile(systemdPath, remoteSystemdUnitFile); err != nil {
							return fmt.Errorf("reading remote file %s on machine %s failed: %s", systemdPath, machine.ID(), err)
						}
						return nil
					})
					if remoteSystemdUnitFile.String() != systemdUnitFile(machine) {
						return false, nil
					}

					current, ok := currentNodeAgents.Get(machine.ID())
					if ok && current.Commit == orbiterCommit {
						return true, nil
					}

					showVersion := "node-agent --version"

					err = infra.Try(machineMonitor, time.NewTimer(7*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
						var cbErr error
						if response, cbErr = cmp.Execute(nil, showVersion); cbErr != nil {
							return fmt.Errorf("running command %s remotely failed: %w", showVersion, cbErr)
						}
						return nil
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

					healthPath := "/usr/local/bin/health"

					if os.Getenv("MODE") == "DEBUG" {
						// Run node agent in debug mode
						if _, err := machine.Execute(nil, "sudo apt-get update && sudo apt-get install -y git && wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz && sudo tar -zxvf go1.13.3.linux-amd64.tar.gz -C / && sudo chown -R $(id -u):$(id -g) /go && /go/bin/go get -u github.com/go-delve/delve/cmd/dlv && /go/bin/go install github.com/go-delve/delve/cmd/dlv && mv ${HOME}/go/bin/dlv /usr/local/bin"); err != nil {
							panic(err)
						}

						binary = fmt.Sprintf("dlv exec %s --api-version 2 --headless --listen 0.0.0.0:5000 --continue --accept-multiclient --", nodeAgentPath)
					}

					stopSystemd := fmt.Sprintf("sudo systemctl stop %s orbos.health* || true", systemdEntry)
					if err := infra.Try(machineMonitor, time.NewTimer(60*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
						if _, cbErr := cmp.Execute(nil, stopSystemd); cbErr != nil {
							return fmt.Errorf("running command %s remotely failed: %w", stopSystemd, cbErr)
						}
						return nil
					}); err != nil {
						return fmt.Errorf("remotely stopping systemd services failed: %w", err)
					}
					machineMonitor.WithFields(map[string]interface{}{
						"command": stopSystemd,
					}).Debug("Executed command")

					if err := helpers.Fanout([]func() error{
						configure(machine),
						func() error {
							if err := infra.Try(machineMonitor, time.NewTimer(8*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
								if cbErr := cmp.WriteFile(systemdPath, strings.NewReader(systemdUnitFile(machine)), 600); cbErr != nil {
									return fmt.Errorf("creating remote file %s failed: %w", systemdPath, cbErr)
								}
								return nil
							}); err != nil {
								return fmt.Errorf("remotely configuring Node Agent systemd unit failed: %w", err)
							}
							machineMonitor.WithFields(map[string]interface{}{
								"path": systemdPath,
							}).Debug("Written file")
							return nil
						},
						func() error {
							if err := infra.Try(machineMonitor, time.NewTimer(20*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
								if cbErr := cmp.WriteFile(nodeAgentPath, bytes.NewReader(executables.PreBuilt("nodeagent")), 700); cbErr != nil {
									return fmt.Errorf("creating remote file %s failed: %w", nodeAgentPath, cbErr)
								}
								return nil
							}); err != nil {
								return fmt.Errorf("remotely installing Node Agent failed: %w", err)
							}
							machineMonitor.WithFields(map[string]interface{}{
								"path": nodeAgentPath,
							}).Debug("Written file")
							return nil
						},
						func() error {
							if err := infra.Try(machineMonitor, time.NewTimer(20*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
								if cbErr := cmp.WriteFile(healthPath, bytes.NewReader(executables.PreBuilt("health")), 711); cbErr != nil {
									return fmt.Errorf("creating remote file %s failed: %w", healthPath, cbErr)
								}
								return nil
							}); err != nil {
								return fmt.Errorf("remotely installing health executable failed: %w", err)
							}
							machineMonitor.WithFields(map[string]interface{}{
								"path": healthPath,
							}).Debug("Written file")
							return nil
						},
					})(); err != nil {
						return err
					}

					enableSystemd := fmt.Sprintf("sudo systemctl daemon-reload && sudo systemctl enable %s && sudo systemctl restart %s", systemdPath, systemdEntry)
					if err := infra.Try(machineMonitor, time.NewTimer(8*time.Second), 2*time.Second, machine, func(cmp infra.Machine) error {
						if _, cbErr := cmp.Execute(nil, enableSystemd); cbErr != nil {
							return fmt.Errorf("running command %s remotely failed: %w", enableSystemd, cbErr)
						}
						return nil
					}); err != nil {
						return fmt.Errorf("remotely configuring systemd to autostart Node Agent after booting failed: %w", err)
					}
					machineMonitor.WithFields(map[string]interface{}{
						"command": enableSystemd,
					}).Debug("Executed command")

					machineMonitor.Info("Node Agent installed")
					return nil
				}
		}
}
