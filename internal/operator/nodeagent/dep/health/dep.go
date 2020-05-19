package health

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/caos/orbos/internal/operator/nodeagent/dep"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/internal/operator/nodeagent/dep/middleware"
	"github.com/caos/orbos/mntr"
)

const dir string = "/etc/systemd/system/health.wants"

type Installer interface {
	isHealth()
	nodeagent.Installer
}
type healthDep struct {
	monitor mntr.Monitor
	systemd *dep.SystemD
}

func New(monitor mntr.Monitor, systemd *dep.SystemD) Installer {
	return &healthDep{monitor: monitor, systemd: systemd}
}

func (healthDep) isHealth() {}

func (healthDep) Is(other nodeagent.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (healthDep) String() string { return "health" }

func (*healthDep) Equals(other nodeagent.Installer) bool {
	_, ok := other.(*healthDep)
	return ok
}

var r = regexp.MustCompile(`^ExecStart=/usr/local/bin/health --http ([^\s]+) (.*)$`)

func (s *healthDep) Current() (pkg common.Package, err error) {

	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		content, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return pkg, err
		}

		match := r.FindStringSubmatch(string(content))
		if pkg.Config == nil {
			pkg.Config = make(map[string]string)
		}

		if s.systemd.Active(file.Name()) {
			pkg.Config[match[0]] = match[1]
		}
	}
	return pkg, nil
}

func (s *healthDep) Ensure(_ common.Package, ensure common.Package) error {

	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		s.systemd.Disable(file.Name())
	}

	if err := os.RemoveAll(dir); err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 700); err != nil {
		return err
	}

	var i int
	for location, args := range ensure.Config {
		i++
		svc := fmt.Sprintf("health.%d.service", i)
		if err := ioutil.WriteFile(filepath.Join(dir, svc), []byte(fmt.Sprintf(`
[Unit]
Description=Healthchecks Proxy
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/health --http %s %s
Restart=always
MemoryMax=20M
MemoryLimit=20M
RestartSec=10

[Install]
WantedBy=multi-user.target
`, location, args)), 600); err != nil {
			return err
		}

		if err := s.systemd.Enable(svc); err != nil {
			return err
		}
	}

	return nil
}
