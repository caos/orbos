package nginx

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/internal/operator/nodeagent/dep"
	"github.com/caos/orbos/internal/operator/nodeagent/dep/middleware"
	"github.com/caos/orbos/internal/operator/nodeagent/dep/selinux"
	"github.com/caos/orbos/mntr"
)

const LimitNoFileKey = "LimitNOFILE="
const LimitNoFile8192Entry = LimitNoFileKey + "8192  "

type Installer interface {
	isNgninx()
	nodeagent.Installer
}
type nginxDep struct {
	manager    *dep.PackageManager
	systemd    *dep.SystemD
	monitor    mntr.Monitor
	os         dep.OperatingSystem
	normalizer *regexp.Regexp
}

func New(monitor mntr.Monitor, manager *dep.PackageManager, systemd *dep.SystemD, os dep.OperatingSystem) Installer {
	return &nginxDep{manager, systemd, monitor, os, regexp.MustCompile(`\d+\.\d+\.\d+`)}
}

func (nginxDep) isNgninx() {}

func (nginxDep) Is(other nodeagent.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (nginxDep) String() string { return "NGINX" }

func (*nginxDep) Equals(other nodeagent.Installer) bool {
	_, ok := other.(*nginxDep)
	return ok
}

func (s *nginxDep) Current() (pkg common.Package, err error) {
	if !s.systemd.Active("nginx") {
		return pkg, err
	}

	installed := s.manager.CurrentVersions("nginx")
	if len(installed) == 0 {
		return pkg, nil
	}
	pkg.Version = "v" + s.normalizer.FindString(installed[0].Version)

	config, err := ioutil.ReadFile("/etc/nginx/nginx.conf")
	if err != nil {
		if os.IsNotExist(err) {
			return pkg, nil
		}
		return pkg, err
	}

	pkg.Config = map[string]string{
		"nginx.conf": string(config),
	}

	unitPath, err := s.systemd.UnitPath("nginx")
	if err != nil {
		return pkg, err
	}

	svc, err := ioutil.ReadFile(unitPath)
	if err != nil {
		return pkg, err
	}

	// make pkg config different, so (*nginxDep).Ensure() is called
	if !strings.Contains(string(svc), LimitNoFile8192Entry) {
		pkg.Config["ensuresystemdconf"] = "yes"
	}

	return pkg, nil
}

func (s *nginxDep) Ensure(remove common.Package, ensure common.Package) error {

	if err := selinux.EnsurePermissive(s.monitor, s.os, remove); err != nil {
		return err
	}

	ensureCfg, ok := ensure.Config["nginx.conf"]
	if !ok {
		s.systemd.Disable("nginx")
		os.Remove("/etc/nginx/nginx.conf")
		return nil
	}

	// TODO: I think this should be removed, as this prevents updating nginx
	if _, ok := remove.Config["nginx.conf"]; !ok {

		try := func() error {
			return s.manager.Install(&dep.Software{
				Package: "nginx",
				Version: strings.TrimLeft(ensure.Version, "v"),
			})
		}

		if err := try(); err != nil {

			swmonitor := s.monitor.WithField("software", "NGINX")
			swmonitor.Error(fmt.Errorf("installing software from existing repo failed, trying again after adding repo: %w", err))

			repoURL := "http://nginx.org/packages/centos/$releasever/$basearch/"
			if err := ioutil.WriteFile("/etc/yum.repos.d/nginx.repo", []byte(fmt.Sprintf(`[nginx-stable]
name=nginx stable repo
baseurl=%s
gpgcheck=1
enabled=1
gpgkey=https://nginx.org/keys/nginx_signing.key
module_hotfixes=true`, repoURL)), 0600); err != nil {
				return err
			}

			swmonitor.WithField("url", repoURL).Info("repo added")

			if err := try(); err != nil {
				swmonitor.Error(fmt.Errorf("installing software from %s failed: %w", repoURL, err))
				return err
			}
		}

		if err := os.MkdirAll("/etc/nginx", 0700); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile("/etc/nginx/nginx.conf", []byte(ensureCfg), 0600); err != nil {
		return err
	}

	unitPath, err := s.systemd.UnitPath("nginx")
	if err != nil {
		return err
	}

	if err := dep.ManipulateFile(unitPath, []string{LimitNoFileKey}, nil, func(line string) *string {
		serviceLine := "[Service]"
		if strings.HasPrefix(line, serviceLine) {
			return strPtr(serviceLine + "\n" + LimitNoFile8192Entry)
		}
		return strPtr(line)
	}); err != nil {
		return err
	}

	if err := s.systemd.Enable("nginx"); err != nil {
		return err
	}

	return s.systemd.Start("nginx")
}

func strPtr(str string) *string { return &str }
