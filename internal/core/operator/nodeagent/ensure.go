package nodeagent

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caos/orbiter/internal/core/operator/common"
	"github.com/caos/orbiter/logging"
)

func init() {
	if err := os.MkdirAll("/var/orbiter", 0700); err != nil {
		panic(err)
	}
}

type FirewallEnsurer interface {
	Ensure(common.Firewall) error
}

type FirewallEnsurerFunc func(common.Firewall) error

func (f FirewallEnsurerFunc) Ensure(fw common.Firewall) error {
	return f(fw)
}

type Dependency struct {
	Installer Installer
	Desired   common.Package
	Current   common.Package
	reboot    bool
}

type Converter interface {
	ToDependencies(common.Software) []*Dependency
	ToSoftware([]*Dependency) common.Software
}

type Installer interface {
	Current() (common.Package, error)
	Ensure(uninstall common.Package, install common.Package) (bool, error)
	Equals(other Installer) bool
	Is(other Installer) bool
	fmt.Stringer
}

func ensure(logger logging.Logger, commit string, firewallEnsurer FirewallEnsurer, conv Converter, desired common.NodeAgentSpec) (*common.NodeAgentCurrent, bool, error) {

	curr := &common.NodeAgentCurrent{
		Commit:      commit,
		NodeIsReady: isReady(),
	}

	defer persistReadyness(curr.NodeIsReady)

	if err := firewallEnsurer.Ensure(*desired.Firewall); err != nil {
		return nil, false, err
	}
	curr.Open = *desired.Firewall

	installedSw, err := deriveTraverse(installed, conv.ToDependencies(*desired.Software))
	if err != nil {
		return nil, false, err
	}

	curr.Software = conv.ToSoftware(installedSw)

	divergentSw := deriveFilter(divergent, append([]*Dependency(nil), installedSw...))
	if len(divergentSw) == 0 {
		curr.NodeIsReady = true
		return curr, false, nil
	}

	if curr.NodeIsReady {
		logger.Info("Marking node as unready")
		curr.NodeIsReady = false
		return curr, false, nil
	}

	if !desired.ChangesAllowed {
		logger.Info("Changes are not allowed")
		return curr, false, nil
	}
	ensureDep := ensureFunc(logger)
	ensuredSw, err := deriveTraverse(ensureDep, divergentSw)
	if err != nil {
		return nil, false, err
	}

	curr.Software = conv.ToSoftware(merge(installedSw, ensuredSw))
	return curr, anyReboot(ensuredSw), nil
}

func installed(dep *Dependency) (*Dependency, error) {
	version, err := dep.Installer.Current()
	if err != nil {
		return dep, err
	}
	dep.Current = version
	return dep, nil
}

func divergent(dep *Dependency) bool {
	return !dep.Desired.Equals(dep.Current)
}

func ensureFunc(logger logging.Logger) func(dep *Dependency) (*Dependency, error) {
	return func(dep *Dependency) (*Dependency, error) {
		reboot, err := dep.Installer.Ensure(dep.Current, dep.Desired)
		if err != nil {
			return dep, err
		}

		logger.WithFields(map[string]interface{}{
			"dependency": dep.Installer,
			"from":       dep.Current,
			"to":         dep.Desired,
			"reboot":     reboot,
		}).Info("Ensured dependency")

		dep.Current = dep.Desired
		dep.reboot = reboot

		return dep, nil
	}
}

func anyReboot(deps []*Dependency) bool {
	return deriveAny(func(dep *Dependency) bool { return dep.reboot }, deps)
}

func merge(inferior []*Dependency, prior []*Dependency) []*Dependency {
	keep := deriveFilter(func(item *Dependency) bool {
		return !contains(prior, item)
	}, append([]*Dependency(nil), inferior...))
	return append(keep, prior...)
}

func contains(deps []*Dependency, dep *Dependency) bool {
	return deriveAny(func(item *Dependency) bool {
		return is(item, dep)
	}, deps)
}

func is(this *Dependency, that *Dependency) bool {
	return this.Installer.Is(that.Installer)
}

func persistReadyness(ready bool) {
	if ready {
		if err := ioutil.WriteFile("/var/orbiter/ready", nil, 600); err != nil {
			panic(err)
		}
		return
	}
	if err := os.RemoveAll("/var/orbiter/ready"); err != nil && !os.IsNotExist(err) {
		panic(err)
	}
}

func isReady() bool {
	_, err := os.Stat("/var/orbiter/ready")
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	return err == nil
}
