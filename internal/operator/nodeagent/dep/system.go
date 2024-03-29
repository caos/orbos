//go:generate stringer -type=Packages

package dep

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Packages int

const (
	UnknownPkg Packages = iota
	DebianBased
	REMBased
)

type OperatingSystem struct {
	Packages Packages
}

func (o OperatingSystem) String() string {
	var os string
	switch o {
	case Ubuntu:
		os = "Ubuntu"
	case CentOS:
		os = "CentOS"
	}
	return os
}

var (
	UnknownOS = OperatingSystem{}
	Ubuntu    = OperatingSystem{DebianBased}
	CentOS    = OperatingSystem{REMBased}
)

type OperatingSystemMajor struct {
	OperatingSystem OperatingSystem
	Version         string
}

func (o OperatingSystemMajor) String() string {
	var versionName string
	switch o {
	case Bionic:
		versionName = "18.04 LTS Bionic Beaver"
	case CentOS7:
		versionName = "7"
	}

	return fmt.Sprintf("%s %s", o.OperatingSystem, versionName)
}

var (
	Unknown = OperatingSystemMajor{UnknownOS, ""}
	Bionic  = OperatingSystemMajor{Ubuntu, "bionic"}
	CentOS7 = OperatingSystemMajor{CentOS, "7"}

	//	Debian OperatingSystem = OperatingSystem{DebianBased}
)

func GetOperatingSystem() (OperatingSystemMajor, error) {
	buf := new(bytes.Buffer)
	defer buf.Reset()

	hostnamectl := exec.Command("hostnamectl")
	hostnamectl.Stdout = buf

	if err := hostnamectl.Run(); err != nil {
		return Unknown, fmt.Errorf("running hostnamectl in order to get operating system information failed: %w", err)
	}

	var (
		osLine string
		err    error
	)
	for err == nil {
		osLine, err = buf.ReadString('\n')
		if strings.Contains(osLine, "Operating System") {
			break
		}
	}

	if err != nil {
		return Unknown, fmt.Errorf("finding line containing \"Operating System\" in hostnamectl output failed: %w", err)
	}

	os := strings.Fields(osLine)[2]
	switch os {
	case "Ubuntu":
		version := strings.Fields(osLine)[3]
		if strings.HasPrefix(version, "18.04") {
			return Bionic, nil
		}
		return Unknown, fmt.Errorf("unsupported ubuntu version %s", version)
	case "CentOS":
		version := strings.Fields(osLine)[4]
		if version == "7" {
			return CentOS7, nil
		}
		return Unknown, fmt.Errorf("unsupported centOS version %s", version)
	}
	return Unknown, fmt.Errorf("unknown operating system %s", os)
}
