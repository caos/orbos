package gce

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"

	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ssh"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	"google.golang.org/api/compute/v1"
)

var _ infra.Machine = (*instance)(nil)

type machine interface {
	Execute(env map[string]string, stdin io.Reader, cmd string) ([]byte, error)
	WriteFile(path string, data io.Reader, permissions uint16) error
	ReadFile(path string, data io.Writer) error
}

type instance struct {
	mntr.Monitor
	id      string
	ip      string
	url     string
	pool    string
	remove  func() error
	context *context
	start   bool
	machine
}

func newMachine(context *context, monitor mntr.Monitor, id, ip, url, pool string, remove func() error, start bool, machine machine) *instance {
	return &instance{
		Monitor: monitor,
		id:      id,
		ip:      ip,
		url:     url,
		pool:    pool,
		remove:  remove,
		context: context,
		start:   start,
		machine: machine,
	}
}

func (c *instance) Execute(env map[string]string, stdin io.Reader, command string) ([]byte, error) {
	buf, err := c.execute(env, stdin, command)
	defer resetBuffer(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *instance) execute(env map[string]string, stdin io.Reader, command string) (outBuf *bytes.Buffer, err error) {
	outBuf = new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	defer resetBuffer(errBuf)

	gcloud, err := exec.LookPath("gcloud")
	if err != nil {
		return nil, err
	}

	if err := gcloudSession(c.context.desired.JSONKey.Value, gcloud, func(bin string) error {
		cmd := exec.Command(gcloud,
			"compute",
			"ssh",
			"--zone", c.context.desired.Zone,
			c.id,
			"--tunnel-through-iap",
			"--project", c.context.projectID,
			"--command", command,
		)
		cmd.Stdin = stdin
		cmd.Stdout = outBuf
		cmd.Stderr = errBuf
		if runErr := cmd.Run(); runErr != nil {
			return errors.New(errBuf.String())
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return outBuf, nil
}

func (c *instance) Shell(env map[string]string) error {
	errBuf := new(bytes.Buffer)
	defer resetBuffer(errBuf)

	gcloud, err := exec.LookPath("gcloud")
	if err != nil {
		return err
	}

	if err := gcloudSession(c.context.desired.JSONKey.Value, gcloud, func(bin string) error {
		cmd := exec.Command(gcloud,
			"compute",
			"ssh",
			"--zone", c.context.desired.Zone,
			c.id,
			"--tunnel-through-iap",
			"--project", c.context.projectID,
		)
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}); err != nil {
		return err
	}
	return nil
}

func (c *instance) WriteFile(path string, data io.Reader, permissions uint16) error {

	user, err := c.Execute(nil, nil, "whoami")
	if err != nil {
		return err
	}

	mkdir, writeFile := ssh.WriteFileCommands(strings.TrimSpace(string(user)), path, permissions)
	_, err = c.Execute(nil, nil, mkdir)
	if err != nil {
		return err
	}

	_, err = c.Execute(nil, data, writeFile)
	return err
}

func (c *instance) ReadFile(path string, data io.Writer) error {
	buf, err := c.execute(nil, nil, fmt.Sprintf("sudo cat %s", path))
	defer resetBuffer(buf)
	if err != nil {
		return err
	}

	_, err = io.Copy(data, buf)
	return err
}

func (c *instance) ID() string {
	return c.id
}

func (c *instance) IP() string {
	return c.ip
}

func (c *instance) Remove() error {
	return c.remove()
}

type instances []*instance

func (c instances) Len() int           { return len(c) }
func (c instances) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c instances) Less(i, j int) bool { return c[i].ID() < c[j].ID() }

func (i instances) strings(field func(i *instance) string) []string {
	sort.Sort(i)
	l := len(i)
	ret := make([]string, l, l)
	for idx, i := range i {
		ret[idx] = field(i)
	}
	return ret
}

func (i instances) refs() []*compute.InstanceReference {
	sort.Sort(i)
	l := len(i)
	ret := make([]*compute.InstanceReference, l, l)
	for idx, i := range i {
		ret[idx] = &compute.InstanceReference{Instance: i.url}
	}
	return ret
}

func ListMachines(monitor mntr.Monitor, desiredTree *tree.Tree, orbID, providerID string) (map[string]infra.Machine, error) {
	desired, err := parseDesiredV0(desiredTree)
	if err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredTree.Parsed = desired

	ctx, err := buildContext(monitor, &desired.Spec, orbID, providerID, true)
	if err != nil {
		return nil, err
	}

	return core.ListMachines(ctx.machinesService)
}
