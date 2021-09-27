package nginx

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/caos/orbos/v5/internal/operator/common"
	"github.com/caos/orbos/v5/internal/operator/nodeagent"
	"github.com/caos/orbos/v5/internal/operator/nodeagent/dep"
	"github.com/caos/orbos/v5/internal/operator/nodeagent/dep/middleware"
	"github.com/caos/orbos/v5/internal/operator/nodeagent/dep/selinux"
	"github.com/caos/orbos/v5/mntr"
)

const LineAddedComment = "# The following line was added by CAOS node agent"
const CleanupLine = "# Line added by CAOS node agent"

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

	systemdUnit, err := ioutil.ReadFile(unitPath)
	if err != nil {
		return pkg, err
	}

	CurrentSystemdEntries(bytes.NewReader(systemdUnit), &pkg)

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

	if err := UpdateSystemdUnitFile(unitPath, ensure.Config); err != nil {
		return err
	}

	if err := s.systemd.Enable("nginx"); err != nil {
		return err
	}

	return s.systemd.Start("nginx")
}

func CurrentSystemdEntries(r io.Reader, p *common.Package) {

	sectionRegexp := regexp.MustCompile("^\\[[a-zA-Z]+]$")
	scanner := bufio.NewScanner(r)
	var addNextLine bool
	var currentSection string
	for scanner.Scan() {
		line := scanner.Text()
		section := sectionRegexp.FindString(line)
		if len(section) >= 1 {
			currentSection = section
		}

		lineparts := strings.Split(line, "=")

		if strings.Contains(line, CleanupLine) {
			p.Config["ensuresystemdconf"] = "yes"
		}

		if addNextLine {
			p.Config[fmt.Sprintf("Systemd%s%s", currentSection, lineparts[0])] = lineparts[1]
			addNextLine = false
		}
		addNextLine = strings.Contains(line, LineAddedComment)
	}
}

func UpdateSystemdUnitFile(path string, cfg map[string]string) error {

	removeContaining := []string{CleanupLine, LineAddedComment}
	sectionRegexpStr := "\\[[a-zA-Z]+\\]"
	keyPartsRegexp := regexp.MustCompile(fmt.Sprintf("^Systemd(%s)([a-zA-Z]+)$", sectionRegexpStr))
	sectionRegexpLine := regexp.MustCompile(fmt.Sprintf("^%s$", sectionRegexpStr))

	for k := range cfg {
		parts := keyPartsRegexp.FindStringSubmatch(k)
		if len(parts) == 3 {
			removeContaining = append(removeContaining, parts[2]+"=")
		}
	}

	return dep.ManipulateFile(path, removeContaining, nil, func(line string) *string {

		if !sectionRegexpLine.MatchString(line) {
			return strPtr(line)
		}

		addLines := []string{line}

		for k, v := range cfg {
			parts := keyPartsRegexp.FindStringSubmatch(k)
			if len(parts) == 3 {
				if parts[1] == line {
					addLines = append(addLines, LineAddedComment, fmt.Sprintf("%s=%s", parts[2], v))
				}
			}
		}

		return strPtr(strings.Join(addLines, "\n"))
	})
}

func strPtr(str string) *string { return &str }
