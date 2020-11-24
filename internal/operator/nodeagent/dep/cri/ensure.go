package cri

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/nodeagent/dep"
)

func (c *criDep) ensureCentOS(runtime string, version string) error {
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	cmd := exec.Command("yum", "remove", "docker",
		"docker-client",
		"docker-client-latest",
		"docker-common",
		"docker-latest",
		"docker-latest-logrotate",
		"docker-logrotate",
		"docker-engine")
	cmd.Stderr = errBuf
	if c.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "removing older docker versions failed with stderr %s", errBuf.String())
	}

	for _, pkg := range []string{"device-mapper-persistent-data", "lvm2"} {
		if err := c.manager.Install(&dep.Software{Package: pkg}); err != nil {
			c.monitor.Error(errors.Wrap(err, "installing docker dependency failed"))
		}
	}

	if err := c.manager.Install(&dep.Software{
		Package: "containerd.io",
	}); err != nil {
		c.monitor.Error(err)
	}

	if err := c.manager.Install(&dep.Software{
		Package: runtime,
		Version: version,
	}); err != nil {
		c.monitor.Error(errors.Wrap(err, "installing container runtime failed"))
	}

	c.manager.Add(&dep.Repository{
		Repository: "https://download.docker.com/linux/centos/docker-ce.repo",
	})

	if err := c.systemd.Enable("docker"); err != nil {
		return err
	}
	return c.systemd.Start("docker")
}

func (c *criDep) ensureUbuntu(runtime string, version string) error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	buf := new(bytes.Buffer)
	defer buf.Reset()

	var versionLine string
	cmd := exec.Command("apt-cache", "madison", runtime)
	cmd.Stderr = errBuf
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "running apt-cache madison %s failed with stderr %s", runtime, errBuf.String())
	}
	errBuf.Reset()

	if c.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
	}

	var err error
	for err == nil {
		versionLine, err = buf.ReadString('\n')
		if c.monitor.IsVerbose() {
			fmt.Println(versionLine)
		}
		if strings.Contains(versionLine, version) {
			break
		}
	}
	buf.Reset()

	if err != nil && versionLine == "" {
		return errors.Wrapf(err, "finding line containing desired container runtime version \"%s\" failed", version)
	}

	if err := c.manager.Install(&dep.Software{Package: "containerd.io"}); err != nil {
		return err
	}

	if err := c.manager.Install(&dep.Software{
		Package: runtime,
		Version: strings.TrimSpace(strings.Split(versionLine, "|")[1]),
	}); err != nil {
		c.monitor.Error(errors.Wrap(err, "installing container runtime failed"))
	}

	c.manager.Add(&dep.Repository{
		Repository:     fmt.Sprintf("deb [arch=amd64] https://download.docker.com/linux/ubuntu %s stable", c.os.Version),
		KeyURL:         "https://download.docker.com/linux/ubuntu/gpg",
		KeyFingerprint: "0EBFCD88",
	})

	if err := c.systemd.Enable("docker"); err != nil {
		return err
	}
	return c.systemd.Start("docker")
}
