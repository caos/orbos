package dynamic

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/caos/orbos/internal/operator/nodeagent/dep/sysctl"

	"github.com/caos/orbos/internal/tree"

	"github.com/caos/orbos/internal/helpers"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/mntr"
)

var probes = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "probe",
		Help: "Load Balancing Probes.",
	},
	[]string{"name", "type", "target"},
)

func init() {
	prometheus.MustRegister(probes)
}

type WhiteListFunc func() []*orbiter.CIDR

func AdaptFunc(whitelist WhiteListFunc) orbiter.AdaptFunc {

	return func(monitor mntr.Monitor, finishedChan chan bool, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, migrate bool, err error) {

		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		if desiredTree.Common.Version != "v2" {
			migrate = true
		}
		desiredKind := &Desired{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, migrate, errors.Wrapf(err, "unmarshaling desired state for kind %s failed", desiredTree.Common.Kind)
		}

		for _, pool := range desiredKind.Spec {
			for _, vip := range pool {
				for _, t := range vip.Transport {
					sort.Strings(t.BackendPools)
					if t.Name == "kubeapi" {
						if t.HealthChecks.Path != "/healthz" {
							t.HealthChecks.Path = "/healthz"
							migrate = true
						}
						if t.HealthChecks.Protocol != "https" {
							t.HealthChecks.Protocol = "https"
							migrate = true
						}
					}
					if len(t.Whitelist) == 0 {
						allIPs := orbiter.CIDR("0.0.0.0/0")
						t.Whitelist = []*orbiter.CIDR{&allIPs}
						migrate = true
					}
				}
			}
		}

		if err := desiredKind.Validate(); err != nil {
			return nil, nil, migrate, err
		}
		desiredTree.Parsed = desiredKind

		current := &Current{
			Common: &tree.Common{
				Kind:    "orbiter.caos.ch/DynamicLoadBalancer",
				Version: "v0",
			},
		}
		currentTree.Parsed = current

		return func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeagents map[string]*common.NodeAgentSpec, queried map[string]interface{}) (orbiter.EnsureFunc, error) {

			wl := whitelist()

			addresses := make(map[string]*infra.Address)
			for _, pool := range desiredKind.Spec {
				for _, vip := range pool {
					for _, t := range vip.Transport {
						addresses[t.Name] = &infra.Address{
							Location:     vip.IP,
							FrontendPort: uint16(t.FrontendPort),
							BackendPort:  uint16(t.BackendPort),
						}
					}
				}
			}

			current.Current.Spec = desiredKind.Spec
			current.Current.Desire = func(forPool string, svc core.MachinesService, balanceLoad bool, notifyMaster func(machine infra.Machine, peers infra.Machines, vips []*VIP) string, mapVIP func(*VIP) string) (bool, error) {

				done := true
				desireNodeAgent := func(machine infra.Machine, fw common.Firewall, nginx, keepalived common.Package) {
					machineMonitor := monitor.WithField("machine", machine.ID())
					deepNa, ok := nodeagents[machine.ID()]
					if !ok {
						deepNa = &common.NodeAgentSpec{}
						nodeagents[machine.ID()] = deepNa
					}
					deepNaCurr := common.NodeAgentCurrent{}
					if naCurr, ok := nodeAgentsCurrent[machine.ID()]; ok {
						deepNaCurr = *naCurr
					}
					if deepNa.Firewall == nil {
						deepNa.Firewall = &common.Firewall{}
					}

					if !deepNa.Firewall.Contains(fw) {
						deepNa.Firewall.Merge(fw)
						machineMonitor.Changed("Loadbalancing Firewall desired")
					}
					if !fw.IsContainedIn(deepNaCurr.Open) {
						machineMonitor.WithField("ports", deepNa.Firewall.Ports()).Info("Awaiting firewalld config")
						done = false
					}
					for _, port := range fw.Ports() {
						if portInt, parseErr := strconv.ParseInt(port.Port, 10, 16); parseErr == nil && portInt == 22 {

							if deepNa.Software.SSHD.Config == nil || deepNa.Software.SSHD.Config["listenaddress"] != machine.IP() {
								deepNa.Software.SSHD.Config = map[string]string{"listenaddress": machine.IP()}
								machineMonitor.Changed("sshd config desired")
							}

							if !deepNaCurr.Software.SSHD.Equals(deepNa.Software.SSHD) {
								machineMonitor.Info("Awaiting sshd config")
								done = false
							}
						}
					}
					if deepNa.Software == nil {
						deepNa.Software = &common.Software{}
					}

					if !nginx.Equals(common.Package{}) {
						if !deepNa.Software.Nginx.Equals(nginx) {
							deepNa.Software.Nginx = nginx
							machineMonitor.Changed("NGINX desired")
						}
						if !deepNa.Software.Nginx.Equals(deepNaCurr.Software.Nginx) {
							machineMonitor.Info("Awaiting NGINX")
							done = false
						}
						if !sysctl.Contains(deepNa.Software.Sysctl, common.Package{
							Config: map[string]string{
								string(sysctl.IpForward):    "1",
								string(sysctl.NonLocalBind): "1",
							},
						}) {
							sysctl.Enable(&deepNa.Software.Sysctl, sysctl.IpForward)
							sysctl.Enable(&deepNa.Software.Sysctl, sysctl.NonLocalBind)
							machineMonitor.Changed("sysctl desired")
						}
						if !sysctl.Contains(deepNaCurr.Software.Sysctl, deepNa.Software.Sysctl) {
							machineMonitor.Info("Awaiting sysctl config")
							done = false
						}
					}

					if !keepalived.Equals(common.Package{}) {
						if !deepNa.Software.KeepaliveD.Equals(keepalived) {
							deepNa.Software.KeepaliveD = keepalived
							machineMonitor.Changed("Keepalived desired")
						}
						deepNa.Software.KeepaliveD = keepalived
						if !deepNa.Software.KeepaliveD.Equals(deepNaCurr.Software.KeepaliveD) {
							monitor.WithField("package", deepNa.Software.KeepaliveD).Info("Awaiting keepalived")
							done = false
						}
					}
				}

				templateFuncs := template.FuncMap(map[string]interface{}{
					"forMachines": svc.List,
					"add": func(i, y int) int {
						return i + y
					},
					"user": func(machine infra.Machine) (string, error) {
						var user string
						whoami := "whoami"
						stdout, err := machine.Execute(nil, nil, whoami)
						if err != nil {
							return "", errors.Wrapf(err, "running command %s remotely failed", whoami)
						}
						user = strings.TrimSuffix(string(stdout), "\n")
						monitor.WithFields(map[string]interface{}{
							"user":    user,
							"machine": machine.ID(),
							"command": whoami,
						}).Debug("Executed command")
						return user, nil
					},
					"vip": mapVIP,
				})

				var nginxNATTemplate *template.Template
				var vips []*VIP
				var lbMachines []infra.Machine

				if !balanceLoad {
					for _, desiredVIPs := range desiredKind.Spec {
						vips = append(vips, desiredVIPs...)
					}
					nginxNATTemplate = template.Must(template.New("").Funcs(templateFuncs).Parse(`events {
	worker_connections  4096;  ## Default: 1024
}

stream { {{ range $nat := .NATs }}
	upstream {{ $nat.Name }} {
		server {{ $nat.To }};
	}

{{ range $from := $nat.From }}	server {
		listen {{ $from }};
		proxy_pass {{ $nat.Name }};
	}
{{ end }}{{ end }}}

`))

				} else {

					var ok bool
					vips, ok = desiredKind.Spec[forPool]
					if !ok {
						return false, nil
					}

					allPools, err := svc.ListPools()
					if err != nil {
						return false, err
					}

					for _, pool := range allPools {
						machines, err := svc.List(pool)
						if err != nil {
							return false, err
						}
						if forPool == pool {
							lbMachines = machines
						}
						for _, machine := range machines {
							cidr := orbiter.CIDR(fmt.Sprintf("%s/32", machine.IP()))
							vips = addToWhitelists(false, vips, &cidr)
						}
					}

					vips = addToWhitelists(true, vips, wl...)

					lbData := make([]LB, len(lbMachines))
					for idx, machine := range lbMachines {
						lbData[idx] = LB{
							VIPs: vips,
							Self: machine,
							Peers: deriveFilterMachines(func(cmp infra.Machine) bool {
								return cmp.ID() != machine.ID()
							}, append([]infra.Machine(nil), lbMachines...)),
							State:                "BACKUP",
							CustomMasterNotifyer: notifyMaster != nil,
						}
						if idx == 0 {
							lbData[idx].State = "MASTER"
						}
					}

					keepaliveDTemplate := template.Must(template.New("").Funcs(templateFuncs).Parse(`{{ $root := . }}global_defs {
	enable_script_security
	script_user {{ user $root.Self }}
}

vrrp_sync_group VG1 {
	group {
{{ range $idx, $_ := .VIPs }}        VI_{{ $idx }}
{{ end }}    }
}

{{ range $idx, $vip := .VIPs }}vrrp_script chk_{{ vip $vip }} {
	script       "/usr/local/bin/health 200@http://127.0.0.1:29999/ready"
	interval 2   # check every 2 seconds
	fall 15      # require 15 failures for KO
	rise 2       # require 2 successes for OK
}

vrrp_instance VI_{{ $idx }} {
	state {{ $root.State }}
	unicast_src_ip {{ $root.Self.IP }}
	unicast_peer {
		{{ range $peer := $root.Peers }}{{ $peer.IP }}
		{{ end }}    }
	interface eth0
	virtual_router_id {{ add 55 $idx }}
	advert_int 1
	authentication {
		auth_type PASS
		auth_pass [ REDACTED ]
	}
	track_script {
		chk_{{ vip $vip }}
	}

{{ if $root.CustomMasterNotifyer }}	notify_master "/etc/keepalived/notifymaster.sh {{ $root.Self.ID }} {{ vip $vip }}"
{{ else }}	virtual_ipaddress {
		{{ vip $vip }}
	}
{{ end }}
}
{{ end }}
`))

					nginxLBTemplate := template.Must(template.New("").Funcs(templateFuncs).Parse(`{{ $root := . }}events {
	worker_connections  4096;  ## Default: 1024
}

stream { {{ range $vip := .VIPs }}{{ range $src := $vip.Transport }}
	upstream {{ $src.Name }} {    {{ range $dest := $src.BackendPools }}{{ range $machine := forMachines $dest }}
		server {{ $machine.IP }}:{{ $src.BackendPort }}; # {{ $dest }}{{end}}{{ end }}
	}
	server {
		listen {{ vip $vip }}:{{ $src.FrontendPort }};
{{ range $white := $src.Whitelist }}		allow {{ $white }};
{{ end }}
		deny all;
		proxy_pass {{ $src.Name }};
	}
{{ end }}{{ end }}}

http {
	server {
		listen 29999;

		location /ready {
			return 200;
		}
	}
}
`))

					for _, d := range lbData {

						ngxBuf := new(bytes.Buffer)
						//noinspection GoDeferInLoop
						defer ngxBuf.Reset()
						if balanceLoad {
							kaBuf := new(bytes.Buffer)
							defer kaBuf.Reset()

							if err := keepaliveDTemplate.Execute(kaBuf, d); err != nil {
								return false, err
							}

							kaPkg := common.Package{Config: map[string]string{"keepalived.conf": kaBuf.String()}}
							kaBuf.Reset()

							if d.CustomMasterNotifyer {
								kaPkg.Config["notifymaster.sh"] = notifyMaster(d.Self, d.Peers, d.VIPs)
							}

							if err := nginxLBTemplate.Execute(ngxBuf, d); err != nil {
								return false, err
							}
							ngxPkg := common.Package{Config: map[string]string{"nginx.conf": ngxBuf.String()}}
							ngxBuf.Reset()

							desireNodeAgent(d.Self, common.Firewall{}, ngxPkg, kaPkg)
						}
					}
				}

				nodesNats := make(map[string]*NATDesires)
				for _, vip := range vips {
					for _, transport := range vip.Transport {
						srcFW := map[string]*common.Allowed{
							fmt.Sprintf("%s-%d-src", transport.Name, transport.FrontendPort): {
								Port:     fmt.Sprintf("%d", transport.FrontendPort),
								Protocol: "tcp",
							},
						}
						probeVIP := func() {
							probe("VIP", vip.IP, uint16(transport.FrontendPort), transport.HealthChecks, *transport)
						}

						var natVIPProbed bool
						if balanceLoad {
							for _, machine := range lbMachines {
								desireNodeAgent(machine, srcFW, common.Package{}, common.Package{})
							}
							probeVIP()
						}
						for _, dest := range transport.BackendPools {

							destFW := map[string]*common.Allowed{
								fmt.Sprintf("%s-%d-dest", transport.Name, transport.BackendPort): {
									Port:     fmt.Sprintf("%d", transport.BackendPort),
									Protocol: "tcp",
								},
							}

							destMachines, err := svc.List(dest)
							if err != nil {
								return false, err
							}

							for _, machine := range destMachines {
								desireNodeAgent(machine, destFW, common.Package{}, common.Package{})
								probe("Upstream", machine.IP(), uint16(transport.BackendPort), transport.HealthChecks, *transport)
								if !balanceLoad && forPool == dest {
									if !natVIPProbed {
										probeVIP()
										natVIPProbed = true
									}

									nodeNatDesires, ok := nodesNats[machine.IP()]
									if !ok {
										nodeNatDesires = &NATDesires{NATs: make([]*NAT, 0)}
									}
									nodeNatDesires.Firewall = srcFW
									nodeNatDesires.Machine = machine
									nodeNatDesires.NATs = append(nodeNatDesires.NATs, &NAT{
										Name: transport.Name,
										From: []string{
											fmt.Sprintf("%s:%d", mapVIP(vip), transport.FrontendPort),  // VIP
											fmt.Sprintf("%s:%d", machine.IP(), transport.FrontendPort), // Node IP
										},
										To: fmt.Sprintf("%s:%d", machine.IP(), transport.BackendPort),
									})
									nodesNats[machine.IP()] = nodeNatDesires
								}
							}
						}
					}
				}

				for _, node := range nodesNats {
					ngxBuf := new(bytes.Buffer)
					//noinspection GoDeferInLoop
					defer ngxBuf.Reset()
					if err := nginxNATTemplate.Execute(ngxBuf, struct {
						NATs []*NAT
					}{
						NATs: node.NATs,
					}); err != nil {
						return false, err
					}
					ngxPkg := common.Package{Config: map[string]string{"nginx.conf": ngxBuf.String()}}
					ngxBuf.Reset()
					desireNodeAgent(node.Machine, node.Firewall, ngxPkg, common.Package{})
				}

				return done, nil
			}
			return nil, nil
		}, nil, migrate, nil
	}
}

func addToWhitelists(makeUnique bool, vips []*VIP, cidr ...*orbiter.CIDR) []*VIP {
	newVIPs := make([]*VIP, len(vips))
	for vipIdx, vip := range vips {
		newTransport := make([]*Transport, len(vip.Transport))
		for srcIdx, src := range vip.Transport {
			newSource := &Transport{
				Name:         src.Name,
				FrontendPort: src.FrontendPort,
				BackendPort:  src.BackendPort,
				BackendPools: src.BackendPools,
				Whitelist:    append(src.Whitelist, cidr...),
				HealthChecks: src.HealthChecks,
			}
			if makeUnique {
				newSource.Whitelist = unique(src.Whitelist)
			}
			newTransport[srcIdx] = newSource
		}
		newVIPs[vipIdx] = &VIP{
			IP:        vip.IP,
			Transport: newTransport,
		}
	}
	return newVIPs
}

func probe(probeType, ip string, port uint16, hc HealthChecks, source Transport) {
	vipProbe := fmt.Sprintf("%s://%s:%d%s", hc.Protocol, ip, port, hc.Path)
	_, err := helpers.Check(vipProbe, int(hc.Code))
	var success float64
	if err == nil {
		success = 1
	}
	probes.With(prometheus.Labels{
		"name":   source.Name,
		"type":   probeType,
		"target": vipProbe,
	}).Set(success)
}

type NATDesires struct {
	NATs     []*NAT
	Machine  infra.Machine
	Firewall common.Firewall
}

type NAT struct {
	Name string
	From []string
	To   string
}

type LB struct {
	VIPs                 []*VIP
	State                string
	RouterID             int
	Self                 infra.Machine
	Peers                []infra.Machine
	CustomMasterNotifyer bool
}

func unique(s []*orbiter.CIDR) []*orbiter.CIDR {
	m := make(map[string]bool, len(s))
	us := make([]*orbiter.CIDR, len(m))
	for _, elem := range s {
		if len(*elem) != 0 {
			if !m[string(*elem)] {
				us = append(us, elem)
				m[string(*elem)] = true
			}
		}
	}

	cidrs := orbiter.CIDRs(us)
	sort.Sort(cidrs)
	return cidrs
}
