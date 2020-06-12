package gce

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ssh"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
)

var _ infra.Machine = (*instance)(nil)

type instance struct {
	mntr.Monitor
	id      string
	ip      string
	url     string
	pool    string
	remove  func() error
	context *context
	start   bool
}

func newMachine(context *context, monitor mntr.Monitor, id, ip, url, pool string, remove func() error, start bool) *instance {
	return &instance{
		Monitor: monitor,
		id:      id,
		ip:      ip,
		url:     url,
		pool:    pool,
		remove:  remove,
		context: context,
		start:   start,
	}
}

func resetBuffer(buffer *bytes.Buffer) {
	if buffer != nil {
		buffer.Reset()
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

	if err := gcloudSession(c.context, func(bin string) error {
		cmd := exec.Command(gcloudBin(),
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
