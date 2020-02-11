package cri

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/nodeagent/dep"
)

func (c *criDep) ensureCentOS(runtime string, version string) error {

	var errBuf bytes.Buffer
	cmd := exec.Command("yum", "remove", "docker",
		"docker-client",
		"docker-client-latest",
		"docker-common",
		"docker-latest",
		"docker-latest-logrotate",
		"docker-logrotate",
		"docker-engine")
	cmd.Stderr = &errBuf
	if c.logger.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "removing older docker versions failed with stderr %s", errBuf.String())
	}

	for _, pkg := range []*dep.Software{
		&dep.Software{Package: "device-mapper-persistent-data"},
		&dep.Software{Package: "lvm2"},
	} {
		if err := c.manager.Install(pkg); err != nil {
			return errors.Wrap(err, "installing docker dependency failed")
		}
	}

	if err := c.manager.Add(&dep.Repository{
		Repository: "https://download.docker.com/linux/centos/docker-ce.repo",
	}); err != nil {
		return errors.Wrap(err, "adding docker repository failed")
	}

	if err := c.manager.Install(&dep.Software{
		Package: "containerd.io",
	}); err != nil {
		return err
	}

	if err := c.manager.Install(&dep.Software{
		Package: runtime,
		Version: version,
	}); err != nil {
		return errors.Wrap(err, "installing container runtime failed")
	}

	if err := c.systemd.Enable("docker"); err != nil {
		return err
	}
	return c.systemd.Start("docker")
}

func (c *criDep) ensureUbuntu(runtime string, version string) error {

	if err := c.manager.Add(&dep.Repository{
		Repository:     fmt.Sprintf("deb [arch=amd64] https://download.docker.com/linux/ubuntu %s stable", c.os.Version),
		KeyURL:         "https://download.docker.com/linux/ubuntu/gpg",
		KeyFingerprint: "0EBFCD88",
	}); err != nil {
		return errors.Wrap(err, "updating repository indices before installing docker-ce failed")
	}

	var (
		buf    bytes.Buffer
		errBuf bytes.Buffer
	)

	var versionLine string
	cmd := exec.Command("apt-cache", "madison", runtime)
	cmd.Stderr = &errBuf
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "running apt-cache madison %s failed with stderr %s", runtime, errBuf.String())
	}
	errBuf.Reset()

	if c.logger.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
	}

	var err error
	for err == nil {
		versionLine, err = buf.ReadString('\n')
		if c.logger.IsVerbose() {
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
		return errors.Wrap(err, "installing container runtime failed")
	}

	if err := c.systemd.Enable("docker"); err != nil {
		return err
	}
	return c.systemd.Start("docker")
}
