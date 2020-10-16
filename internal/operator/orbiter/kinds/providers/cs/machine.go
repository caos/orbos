package cs

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/cloudscale-ch/cloudscale-go-sdk"
)

var _ infra.Machine = (*machine)(nil)

type action struct {
	required  bool
	require   func()
	unrequire func()
}

type machine struct {
	server *cloudscale.Server
	ip     string
	*ssh.Machine
	remove      func() error
	context     *context
	reboot      *action
	replacement *action
	pool        *Pool
}

func newMachine(server *cloudscale.Server, ip string, sshMachine *ssh.Machine, remove func() error, context *context, pool *Pool) *machine {
	return &machine{
		server:  server,
		ip:      ip,
		Machine: sshMachine,
		remove:  remove,
		context: context,
		pool:    pool,
	}
}

func (m *machine) ID() string    { return m.server.Name }
func (m *machine) IP() string    { return m.ip }
func (m *machine) Remove() error { return m.remove() }

func (m *machine) RebootRequired() (required bool, require func(), unrequire func()) {

	m.reboot = m.initAction(
		m.reboot,
		func() []string { return m.context.desired.RebootRequired },
		func(machines []string) { m.context.desired.RebootRequired = machines })

	return m.reboot.required, m.reboot.require, m.reboot.unrequire
}

func (m *machine) ReplacementRequired() (required bool, require func(), unrequire func()) {

	m.replacement = m.initAction(
		m.replacement,
		func() []string { return m.context.desired.ReplacementRequired },
		func(machines []string) { m.context.desired.ReplacementRequired = machines })

	return m.replacement.required, m.replacement.require, m.replacement.unrequire
}

func (m *machine) initAction(a *action, getSlice func() []string, setSlice func([]string)) *action {
	if a != nil {
		return a
	}

	newAction := &action{
		required:  false,
		unrequire: func() {},
		require: func() {
			s := getSlice()
			s = append(s, m.server.UUID)
			setSlice(s)
		},
	}

	s := getSlice()
	for sIdx := range s {
		req := s[sIdx]
		if req == m.ID() {
			newAction.required = true
			break
		}
	}

	if newAction.required {
		newAction.unrequire = func() {
			s := getSlice()
			for sIdx := range s {
				req := s[sIdx]
				if req == m.ID() {
					s = append(s[0:sIdx], s[sIdx+1:]...)
				}
			}
			setSlice(s)
		}
	}

	return newAction
}
