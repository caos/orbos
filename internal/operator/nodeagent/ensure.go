package nodeagent

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/logging"
)

func init() {
	if err := os.MkdirAll("/var/orbiter", 0700); err != nil {
		panic(err)
	}
}

type FirewallEnsurer interface {
	Ensure(common.Firewall) (bool, error)
}

type FirewallEnsurerFunc func(common.Firewall) (bool, error)

func (f FirewallEnsurerFunc) Ensure(fw common.Firewall) (bool, error) {
	return f(fw)
}

type Dependency struct {
	Installer Installer
	Desired   common.Package
	Current   common.Package
}

type Converter interface {
	ToDependencies(common.Software) []*Dependency
	ToSoftware([]*Dependency) common.Software
}

type Installer interface {
	Current() (common.Package, error)
	Ensure(uninstall common.Package, install common.Package) error
	Equals(other Installer) bool
	Is(other Installer) bool
	fmt.Stringer
}

func ensure(logger logging.Logger, commit string, firewallEnsurer FirewallEnsurer, conv Converter, desired common.NodeAgentSpec, curr *common.NodeAgentCurrent) error {

	curr.Commit = commit
	curr.NodeIsReady = isReady()

	defer persistReadyness(curr.NodeIsReady)

	fwChanged, err := firewallEnsurer.Ensure(*desired.Firewall)
	if err != nil {
		return err
	}
	curr.Open = *desired.Firewall
	if fwChanged {
		logger.Info(true, "Firewall changed")
	}

	installedSw, err := deriveTraverse(installed, conv.ToDependencies(*desired.Software))
	if err != nil {
		return err
	}

	curr.Software = conv.ToSoftware(installedSw)

	divergentSw := deriveFilter(divergent, append([]*Dependency(nil), installedSw...))
	if len(divergentSw) == 0 {
		curr.NodeIsReady = true
		return nil
	}

	if curr.NodeIsReady {
		curr.NodeIsReady = false
		logger.Info(true, "Marked node as unready")
		return nil
	}

	if !desired.ChangesAllowed {
		logger.Info(false, "Changes are not allowed")
		return nil
	}
	ensureDep := ensureFunc(logger)
	ensuredSw, err := deriveTraverse(ensureDep, divergentSw)
	if err != nil {
		return err
	}

	curr.Software = conv.ToSoftware(merge(installedSw, ensuredSw))
	return nil
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
		logger.WithFields(map[string]interface{}{
			"dependency": dep.Installer,
			"from":       dep.Current,
			"to":         dep.Desired,
		}).Info(false, "Ensuring dependency")

		if err := dep.Installer.Ensure(dep.Current, dep.Desired); err != nil {
			return dep, err
		}

		dep.Current = dep.Desired

		logger.WithFields(map[string]interface{}{
			"dependency": dep.Installer,
			"from":       dep.Current,
			"to":         dep.Desired,
		}).Info(true, "Dependency ensured")

		return dep, nil
	}
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
