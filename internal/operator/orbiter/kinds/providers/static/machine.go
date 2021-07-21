package static

import (
	"fmt"
	"strings"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/tree"
)

var _ infra.Machine = (*machine)(nil)

type machine struct {
	poolFile             string
	rebootRequired       bool
	requireReboot        func()
	unrequireReboot      func()
	replacementRequired  bool
	requireReplacement   func()
	unrequireReplacement func()
	*ssh.Machine
	X_ID     *string `header:"id"`
	X_IP     string  `header:"ip"`
	X_active bool    `header:"active"`
}

func newMachine(
	monitor mntr.Monitor,
	poolFile string,
	remoteUser string,
	id *string,
	ip string,
	rebootRequired bool,
	requireReboot func(),
	unrequireReboot func(),
	replacementRequired bool,
	requireReplacement func(),
	unrequireReplacement func(),
) *machine {
	return &machine{
		X_active:             false,
		poolFile:             poolFile,
		X_ID:                 id,
		X_IP:                 ip,
		Machine:              ssh.NewMachine(monitor, remoteUser, ip),
		rebootRequired:       rebootRequired,
		requireReboot:        requireReboot,
		unrequireReboot:      unrequireReboot,
		replacementRequired:  replacementRequired,
		requireReplacement:   requireReplacement,
		unrequireReplacement: unrequireReplacement,
	}
}

func (c *machine) ID() string {
	return *c.X_ID
}

func (c *machine) IP() string {
	return c.X_IP
}

func (c *machine) Remove() error {
	if err := c.Machine.WriteFile(c.poolFile, strings.NewReader(""), 600); err != nil {
		return err
	}
	c.X_active = false
	c.Execute(nil, "sudo systemctl stop node-agentd")
	c.Execute(nil, "sudo systemctl disable node-agentd")
	c.Execute(nil, "sudo kubeadm reset -f")
	c.Execute(nil, "sudo rm -rf /var/lib/etcd")
	return nil
}

func (c *machine) RebootRequired() (bool, func(), func()) {
	return c.rebootRequired, c.requireReboot, c.unrequireReboot
}

func (c *machine) ReplacementRequired() (bool, func(), func()) {
	return c.replacementRequired, c.requireReplacement, c.unrequireReplacement
}

func ListMachines(monitor mntr.Monitor, desiredTree *tree.Tree, providerID string) (map[string]infra.Machine, error) {
	desired, err := parseDesiredV0(desiredTree)
	if err != nil {
		return nil, fmt.Errorf("parsing desired state failed: %w", err)
	}
	desiredTree.Parsed = desired

	machinesSvc := NewMachinesService(monitor,
		desired,
		providerID)

	if err := machinesSvc.updateKeys(); err != nil {
		return nil, err
	}

	return core.ListMachines(machinesSvc)
}
