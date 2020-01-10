package adapter

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caos/orbiter/internal/edge/executables"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/model"
	"github.com/pkg/errors"
)

func installNodeAgent(cfg *model.Config, compute infra.Compute) error {

	var user string
	whoami := "whoami"
	if err := try(cfg.Params.Logger, time.NewTimer(1*time.Minute), 2*time.Second, compute, func(cmp infra.Compute) error {
		var cbErr error
		stdout, cbErr := cmp.Execute(nil, nil, whoami)
		if cbErr != nil {
			return errors.Wrapf(cbErr, "running command %s remotely failed", whoami)
		}
		user = strings.TrimSuffix(string(stdout), "\n")
		return nil
	}); err != nil {
		return errors.Wrap(err, "checking")
	}
	logger := cfg.Params.Logger.WithFields(map[string]interface{}{
		"user":    user,
		"compute": compute.ID(),
	})
	logger.WithFields(map[string]interface{}{
		"command": whoami,
	}).Debug("Executed command")

	dockerCfg := "/etc/docker/daemon.json"
	if err := try(cfg.Params.Logger, time.NewTimer(8*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
		return errors.Wrapf(cmp.WriteFile(dockerCfg, strings.NewReader(`{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
`), 600), "creating remote file %s failed", dockerCfg)
	}); err != nil {
		return errors.Wrap(err, "configuring remote docker failed")
	}
	logger.WithFields(map[string]interface{}{
		"path": dockerCfg,
	}).Debug("Written file")

	systemdEntry := "node-agentd"
	systemdPath := fmt.Sprintf("/lib/systemd/system/%s.service", systemdEntry)

	nodeAgentPath := "/usr/local/bin/node-agent"
	healthPath := "/usr/local/bin/health"

	binary := nodeAgentPath
	if os.Getenv("MODE") == "DEBUG" {
		// Run node agent in debug mode
		if _, err := compute.Execute(nil, nil, "sudo apt-get update && sudo apt-get install -y git && wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz && sudo tar -zxvf go1.13.3.linux-amd64.tar.gz -C / && sudo chown -R $(id -u):$(id -g) /go && /go/bin/go get -u github.com/go-delve/delve/cmd/dlv && /go/bin/go install github.com/go-delve/delve/cmd/dlv && mv ${HOME}/go/bin/dlv /usr/local/bin"); err != nil {
			panic(err)
		}

		binary = fmt.Sprintf("dlv exec %s --api-version 2 --headless --listen 0.0.0.0:5000 --continue --accept-multiclient --", nodeAgentPath)
	}
	if err := try(cfg.Params.Logger, time.NewTimer(8*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
		return errors.Wrapf(cmp.WriteFile(systemdPath, strings.NewReader(fmt.Sprintf(`[Unit]
Description=Node Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=%s --repourl "%s" --id "%s" --currentfile "%s" --secretsfile "%s"
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
`, binary, cfg.Params.RepoURL, compute.ID(), cfg.Params.CurrentFile, cfg.Params.SecretsFile)), 600), "creating remote file %s failed", systemdPath)
	}); err != nil {
		return errors.Wrap(err, "remotely configuring Node Agent systemd unit failed")
	}
	logger.WithFields(map[string]interface{}{
		"path": systemdPath,
	}).Debug("Written file")

	keyPath := "/etc/nodeagent/repokey"
	if err := try(cfg.Params.Logger, time.NewTimer(8*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
		return errors.Wrapf(cmp.WriteFile(keyPath, strings.NewReader(cfg.Params.RepoKey), 400), "creating remote file %s failed", keyPath)
	}); err != nil {
		return errors.Wrap(err, "writing repokey failed")
	}
	logger.WithFields(map[string]interface{}{
		"path": keyPath,
	}).Debug("Written file")

	daemonReload := "sudo systemctl daemon-reload"
	if err := try(cfg.Params.Logger, time.NewTimer(8*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
		_, cbErr := cmp.Execute(nil, nil, daemonReload)
		return errors.Wrapf(cbErr, "running command %s remotely failed", daemonReload)
	}); err != nil {
		return errors.Wrap(err, "reloading remote systemd failed")
	}
	logger.WithFields(map[string]interface{}{
		"command": daemonReload,
	}).Debug("Executed command")

	stopSystemd := fmt.Sprintf("if sudo systemctl is-active %s; then sudo systemctl stop %s;fi", systemdEntry, systemdEntry)
	if err := try(cfg.Params.Logger, time.NewTimer(8*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
		_, cbErr := cmp.Execute(nil, nil, stopSystemd)
		return errors.Wrapf(cbErr, "running command %s remotely failed", stopSystemd)
	}); err != nil {
		return errors.Wrap(err, "remotely stopping Node Agent by systemd failed")
	}
	logger.WithFields(map[string]interface{}{
		"command": stopSystemd,
	}).Debug("Executed command")

	nodeagent, err := executables.PreBuilt("nodeagent")
	if err != nil {
		return err
	}
	if err := try(cfg.Params.Logger, time.NewTimer(20*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
		return errors.Wrapf(cmp.WriteFile(nodeAgentPath, bytes.NewReader(nodeagent), 700), "creating remote file %s failed", nodeAgentPath)
	}); err != nil {
		return errors.Wrap(err, "remotely installing Node Agent failed")
	}
	logger.WithFields(map[string]interface{}{
		"path": nodeAgentPath,
	}).Debug("Written file")

	health, err := executables.PreBuilt("health")
	if err != nil {
		return err
	}
	if err := try(cfg.Params.Logger, time.NewTimer(20*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
		return errors.Wrapf(cmp.WriteFile(healthPath, bytes.NewReader(health), 711), "creating remote file %s failed", healthPath)
	}); err != nil {
		return errors.Wrap(err, "remotely installing health executable failed")
	}
	logger.WithFields(map[string]interface{}{
		"path": healthPath,
	}).Debug("Written file")

	enableSystemd := fmt.Sprintf("sudo systemctl enable %s", systemdPath)
	if err := try(cfg.Params.Logger, time.NewTimer(8*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
		_, cbErr := cmp.Execute(nil, nil, enableSystemd)
		return errors.Wrapf(cbErr, "running command %s remotely failed", enableSystemd)
	}); err != nil {
		return errors.Wrap(err, "remotely configuring systemd to autostart Node Agent after booting failed")
	}
	logger.WithFields(map[string]interface{}{
		"command": enableSystemd,
	}).Debug("Executed command")

	startSystemd := fmt.Sprintf("sudo systemctl restart %s", systemdEntry)
	if err := try(cfg.Params.Logger, time.NewTimer(8*time.Second), 2*time.Second, compute, func(cmp infra.Compute) error {
		_, cbErr := cmp.Execute(nil, nil, startSystemd)
		return errors.Wrapf(cbErr, "running command %s remotely failed", startSystemd)
	}); err != nil {
		return errors.Wrap(err, "remotely starting Node Agent by systemd failed")
	}

	logger.WithFields(map[string]interface{}{
		"command": startSystemd,
	}).Debug("Executed command")

	logger.WithFields(map[string]interface{}{
		"compute": compute.ID(),
	}).Info("Node Agent installed and started")

	return nil

}
