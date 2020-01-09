//go:generate goderive . -dedup -autoname

package adapter

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/nodeagent/model"
	"github.com/caos/orbiter/logging"
)

func init() {
	if err := os.MkdirAll("/var/orbiter", 0700); err != nil {
		panic(err)
	}
}

type FirewallEnsurer interface {
	Ensure(operator.Firewall) error
}

type FirewallEnsurerFunc func(operator.Firewall) error

func (f FirewallEnsurerFunc) Ensure(fw operator.Firewall) error {
	return f(fw)
}

type Rebooter interface {
	Reboot() error
}

type Dependency struct {
	Installer Installer
	Desired   operator.Package
	Current   operator.Package
	reboot    bool
}

type Converter interface {
	ToDependencies(operator.Software) []*Dependency
	ToSoftware([]*Dependency) operator.Software
}

type Installer interface {
	Current() (operator.Package, error)
	Ensure(uninstall operator.Package, install operator.Package) (bool, error)
	Equals(other Installer) bool
	Is(other Installer) bool
	fmt.Stringer
}

type Callback func() error

func New(commit string, logger logging.Logger, rebooter Rebooter, firewallEnsurer FirewallEnsurer, conv Converter, before Callback, after Callback) Builder {
	return builderFunc(func(userSpec model.UserSpec, nodeagentUpdater operator.NodeAgentUpdater) (model.Config, Adapter, error) {
		if userSpec.Verbose && !logger.IsVerbose() {
			logger = logger.Verbose()
		}
		return model.Config{}, adapterFunc(func(ctx context.Context, secrets *operator.Secrets, deps map[string]interface{}) (curr *model.Current, err error) {

			if before != nil {
				if err := before(); err != nil {
					return nil, err
				}
			}

			curr = &model.Current{
				Commit:      commit,
				NodeIsReady: isReady(),
			}

			defer persistReadyness(curr.NodeIsReady)

			if err = firewallEnsurer.Ensure(userSpec.Firewall); err != nil {
				return nil, err
			}
			curr.Open = userSpec.Firewall

			installedSw, err := deriveTraverse(installed, conv.ToDependencies(*userSpec.Software))
			if err != nil {
				return curr, err
			}

			curr.Software = conv.ToSoftware(installedSw)

			divergentSw := deriveFilter(divergent, append([]*Dependency(nil), installedSw...))
			if len(divergentSw) == 0 {
				curr.NodeIsReady = true
				return curr, nil
			}

			if curr.NodeIsReady {
				logger.Info("Marking node as unready")
				curr.NodeIsReady = false
				return curr, nil
			}

			if !userSpec.ChangesAllowed {
				logger.Info("Changes are not allowed")
				return curr, nil
			}
			ensure := ensureFunc(logger)
			ensuredSw, err := deriveTraverse(ensure, divergentSw)
			if err != nil {
				return nil, err
			}

			curr.Software = conv.ToSoftware(merge(installedSw, ensuredSw))

			if anyReboot(ensuredSw) {
				//				if !userSpec.RebootEnabled {
				//					logger.Info("Rebooting is not enabled")
				//					return curr, nil
				//				}
				logger.Info("Rebooting")
				if err = rebooter.Reboot(); err != nil {
					return nil, err
				}
			}

			if after != nil {
				return curr, after()
			}

			return curr, nil
		}), nil
	})
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
