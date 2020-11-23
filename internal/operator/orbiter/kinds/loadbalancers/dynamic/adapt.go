package dynamic

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/caos/orbos/pkg/secret"
	"github.com/pires/go-proxyproto"

	"github.com/caos/orbos/internal/operator/nodeagent/dep/sysctl"

	"github.com/caos/orbos/pkg/tree"

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

type VRRP struct {
	VRRPInterface string
	NotifyMaster  func(machine infra.Machine) (string, bool)
	AuthCheck     func(machine infra.Machine) (string, int)
}

func AdaptFunc(whitelist WhiteListFunc) orbiter.AdaptFunc {
	return func(monitor mntr.Monitor, finishedChan chan struct{}, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, configureFunc orbiter.ConfigureFunc, migrate bool, secrets map[string]*secret.Secret, err error) {

		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		if desiredTree.Common.Version != "v2" {
			migrate = true
		}
		desiredKind := &Desired{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, nil, migrate, nil, errors.Wrapf(err, "unmarshaling desired state for kind %s failed", desiredTree.Common.Kind)
		}

		for _, pool := range desiredKind.Spec {
			for _, vip := range pool {
				for _, t := range vip.Transport {
					sort.Strings(t.BackendPools)
					if t.ProxyProtocol == nil {
						trueVal := true
						t.ProxyProtocol = &trueVal
						migrate = true
					}
					if t.Name == "kubeapi" {
						if t.HealthChecks.Path != "/healthz" {
							t.HealthChecks.Path = "/healthz"
							migrate = true
						}
						if t.HealthChecks.Protocol != "https" {
							t.HealthChecks.Protocol = "https"
							migrate = true
						}
						if t.ProxyProtocol == nil || *t.ProxyProtocol {
							f := false
							t.ProxyProtocol = &f
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
			return nil, nil, nil, migrate, nil, err
		}
		desiredTree.Parsed = desiredKind

		current := &Current{
			Common: &tree.Common{
				Kind:    "orbiter.caos.ch/DynamicLoadBalancer",
				Version: "v0",
			},
		}
		currentTree.Parsed = current

		return func(nodeAgentsCurrent *common.CurrentNodeAgents, nodeagents *common.DesiredNodeAgents, queried map[string]interface{}) (orbiter.EnsureFunc, error) {

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

			poolMachines := curryPoolMachines()
			enrichedVIPs := curryEnrichedVIPs(*desiredKind, poolMachines, wl, nodeAgentsCurrent)

			current.Current.Spec = enrichedVIPs
			current.Current.Desire = func(forPool string, svc core.MachinesService, vrrp *VRRP, mapVIP func(*VIP) string) (bool, error) {

				var lbMachines []infra.Machine

				done := true
				desireNodeAgent := func(machine infra.Machine, fw common.Firewall, nginx, keepalived common.Package) {
					machineMonitor := monitor.WithField("machine", machine.ID())
					deepNa, _ := nodeagents.Get(machine.ID())
					deepNaCurr, _ := nodeAgentsCurrent.Get(machine.ID())

					if !deepNa.Firewall.Contains(fw) {
						deepNa.Firewall.Merge(fw)
						machineMonitor.Changed("Loadbalancing firewall desired")
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
							monitor.Info("Awaiting keepalived")
							done = false
						}
					}
				}

				templateFuncs := template.FuncMap(map[string]interface{}{
					"forMachines": svc.List,
					"add":         func(i, y int) int { return i + y },
					"user": func(machine infra.Machine) (string, error) {
						var user string
						whoami := "whoami"
						stdout, err := machine.Execute(nil, whoami)
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
					"vip":       mapVIP,
					"derefBool": func(in *bool) bool { return in != nil && *in },
				})

				var nginxNATTemplate *template.Template
				var vips []*VIP

				if vrrp == nil {
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

{{ range $white := $nat.Whitelist }}		allow {{ $white }};
{{ end }}
		deny all;
		proxy_pass {{ $nat.Name }};
		proxy_protocol {{ if derefBool $nat.ProxyProtocol }}on{{ else }}off{{ end }};
	}
{{ end }}{{ end }}}`))

				} else {

					if err := poolMachines(svc, func(pool string, machines infra.Machines) {

						if forPool == pool {
							lbMachines = machines
						}
					}); err != nil {
						return false, err
					}

					spec, _, err := enrichedVIPs(svc)
					if err != nil {
						return false, err
					}

					lbData := make([]LB, len(lbMachines))
					for idx, machine := range lbMachines {
						lbData[idx] = LB{
							VIPs: spec[forPool],
							Self: machine,
							Peers: deriveFilterMachines(func(cmp infra.Machine) bool {
								return cmp.ID() != machine.ID()
							}, append([]infra.Machine(nil), lbMachines...)),
							State:                "BACKUP",
							CustomMasterNotifyer: vrrp.NotifyMaster != nil,
							Interface:            vrrp.VRRPInterface,
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
	interface {{ $root.Interface }}
	virtual_router_id {{ add 55 $idx }}
	advert_int 1
	authentication {
		auth_type PASS
		auth_pass [ REDACTED ]
	}
	track_script {
		chk_{{ vip $vip }}
	}

{{ if $root.CustomMasterNotifyer }}	notify_master "/etc/keepalived/notifymaster.sh"
{{ else }}	virtual_ipaddress {
		{{ vip $vip }}
	}
{{ end }}
}
{{ end }}`))

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
		proxy_protocol {{ if derefBool $src.ProxyProtocol }}on{{ else }}off{{ end }};
	}
{{ end }}{{ end }}}

http {
	server {
		listen 29999;

		location /ready {
			return 200;
		}
	}
}`))

					for _, d := range lbData {

						ngxBuf := new(bytes.Buffer)
						//noinspection GoDeferInLoop
						defer ngxBuf.Reset()
						kaBuf := new(bytes.Buffer)
						defer kaBuf.Reset()

						if err := keepaliveDTemplate.Execute(kaBuf, d); err != nil {
							return false, err
						}

						kaPkg := common.Package{Config: map[string]string{"keepalived.conf": kaBuf.String()}}
						kaBuf.Reset()

						if d.CustomMasterNotifyer {
							var enforceEnsuring bool
							kaPkg.Config["notifymaster.sh"], enforceEnsuring = vrrp.NotifyMaster(d.Self)
							if enforceEnsuring {
								kaPkg.Config["reensure"] = "true"
							}
						}

						if vrrp.AuthCheck != nil {
							authCheck, expectedExitCode := vrrp.AuthCheck(d.Self)
							if authCheck != "" {
								kaPkg.Config["authcheck.sh"] = authCheck
								kaPkg.Config["authcheckexitcode"] = strconv.Itoa(expectedExitCode)
							}
						}

						if err := nginxLBTemplate.Execute(ngxBuf, d); err != nil {
							return false, err
						}
						ngxPkg := common.Package{Config: map[string]string{"nginx.conf": ngxBuf.String()}}
						ngxBuf.Reset()

						desireNodeAgent(d.Self, common.ToFirewall(make(map[string]*common.Allowed)), ngxPkg, kaPkg)
					}
				}

				nodesNats := make(map[string]*NATDesires)
				spec, _, err := enrichedVIPs(svc)
				if err != nil {
					return false, err
				}
				for _, vips := range spec {
					for _, vip := range vips {
						for _, transport := range vip.Transport {
							srcFW := map[string]*common.Allowed{
								fmt.Sprintf("%s-%d-src", transport.Name, transport.FrontendPort): {
									Port:     fmt.Sprintf("%d", transport.FrontendPort),
									Protocol: "tcp",
								},
							}
							ip := mapVIP(vip)
							probeVIP := func() {
								probe("VIP", ip, uint16(transport.FrontendPort), false, transport.HealthChecks, *transport)
							}

							var natVIPProbed bool
							if vrrp != nil {
								for _, machine := range lbMachines {
									desireNodeAgent(machine, common.ToFirewall(srcFW), common.Package{}, common.Package{})
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
									desireNodeAgent(machine, common.ToFirewall(destFW), common.Package{}, common.Package{})
									probe("Upstream", machine.IP(), uint16(transport.BackendPort), *transport.ProxyProtocol, transport.HealthChecks, *transport)
									if vrrp == nil && forPool == dest {
										if !natVIPProbed {
											probeVIP()
											natVIPProbed = true
										}

										nodeNatDesires, ok := nodesNats[machine.IP()]
										if !ok {
											nodeNatDesires = &NATDesires{NATs: make([]*NAT, 0)}
										}
										nodeNatDesires.Firewall = common.ToFirewall(srcFW)
										nodeNatDesires.Machine = machine

										nodeNatDesires.NATs = append(nodeNatDesires.NATs, &NAT{
											Whitelist: transport.Whitelist,
											Name:      transport.Name,
											From: []string{
												fmt.Sprintf("%s:%d", ip, transport.FrontendPort),           // VIP
												fmt.Sprintf("%s:%d", machine.IP(), transport.FrontendPort), // Node IP
											},
											To:            fmt.Sprintf("%s:%d", machine.IP(), transport.BackendPort),
											ProxyProtocol: *transport.ProxyProtocol,
										})
										nodesNats[machine.IP()] = nodeNatDesires
									}
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
			return orbiter.NoopEnsure, nil
		}, orbiter.NoopDestroy, orbiter.NoopConfigure, migrate, make(map[string]*secret.Secret, 0), nil
	}
}

func addToWhitelists(makeUnique bool, vips []*VIP, cidr ...*orbiter.CIDR) []*VIP {
	newVIPs := make([]*VIP, len(vips))
	for vipIdx, vip := range vips {
		newTransport := make([]*Transport, len(vip.Transport))
		for srcIdx, src := range vip.Transport {
			newSource := &Transport{
				Name:          src.Name,
				FrontendPort:  src.FrontendPort,
				BackendPort:   src.BackendPort,
				BackendPools:  src.BackendPools,
				Whitelist:     append(src.Whitelist, cidr...),
				HealthChecks:  src.HealthChecks,
				ProxyProtocol: src.ProxyProtocol,
			}
			if makeUnique {
				newSource.Whitelist = unique(newSource.Whitelist)
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

func probe(probeType, ip string, port uint16, proxyProtocol bool, hc HealthChecks, source Transport) {
	vipProbe := fmt.Sprintf("%s://%s:%d%s", hc.Protocol, ip, port, hc.Path)

	var success float64
	if proxyProtocol {
		header := &proxyproto.Header{
			Version:           1,
			Command:           proxyproto.PROXY,
			TransportProtocol: proxyproto.TCPv4,
			SourceAddr: &net.TCPAddr{
				IP:   net.ParseIP("10.1.1.1"),
				Port: 1000,
			},
			DestinationAddr: &net.TCPAddr{
				IP:   net.ParseIP(ip),
				Port: int(port),
			},
		}

		_, err := helpers.CheckProxy(vipProbe, int(hc.Code), header)
		if err == nil {
			success = 1
		}
	} else {
		_, err := helpers.Check(vipProbe, int(hc.Code))
		if err == nil {
			success = 1
		}
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
	Name          string
	Whitelist     []*orbiter.CIDR
	From          []string
	To            string
	ProxyProtocol bool
}

type LB struct {
	VIPs                 []*VIP
	State                string
	RouterID             int
	Self                 infra.Machine
	Peers                []infra.Machine
	CustomMasterNotifyer bool
	Interface            string
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

type poolMachinesFunc func(svc core.MachinesService, do func(string, infra.Machines)) error

func curryPoolMachines() poolMachinesFunc {
	var poolsCache map[string]infra.Machines
	return func(svc core.MachinesService, do func(string, infra.Machines)) error {

		if poolsCache == nil {
			poolsCache = make(map[string]infra.Machines)
			allPools, err := svc.ListPools()
			if err != nil {
				return err
			}

			for _, pool := range allPools {
				machines, err := svc.List(pool)
				if err != nil {
					return err
				}
				poolsCache[pool] = machines
			}
		}

		if poolsCache != nil {
			for pool, machines := range poolsCache {
				do(pool, machines)
			}
		}
		return nil
	}
}

func curryEnrichedVIPs(desired Desired, machines poolMachinesFunc, adaptWhitelist []*orbiter.CIDR, nodeAgents *common.CurrentNodeAgents) func(svc core.MachinesService) (map[string][]*VIP, []AuthCheckResult, error) {
	var enrichVIPsCache map[string][]*VIP
	var authCheckResultsCache []AuthCheckResult
	return func(svc core.MachinesService) (map[string][]*VIP, []AuthCheckResult, error) {
		if enrichVIPsCache != nil && authCheckResultsCache != nil {
			return enrichVIPsCache, authCheckResultsCache, nil
		}
		enrichVIPsCache = make(map[string][]*VIP)
		authCheckResultsCache = make([]AuthCheckResult, 0)

		addedCIDRs := append([]*orbiter.CIDR(nil), adaptWhitelist...)
		if err := machines(svc, func(_ string, machines infra.Machines) {
			for _, machine := range machines {
				na, found := nodeAgents.Get(machine.ID())
				if found {
					cfg := na.Software.KeepaliveD.Config
					if cfg != nil {
						authCheckExitCode, ok := cfg["authcheckexitcode"]
						if ok {
							authCheckExitCodeInt, err := strconv.Atoi(authCheckExitCode)
							if err == nil {
								authCheckResultsCache = append(authCheckResultsCache, AuthCheckResult{
									Machine:  machine,
									ExitCode: authCheckExitCodeInt,
								})
							}
						}
					}
				}
				cidr := orbiter.CIDR(fmt.Sprintf("%s/32", machine.IP()))
				addedCIDRs = append(addedCIDRs, &cidr)
			}
		}); err != nil {
			return nil, nil, err
		}
		for deployPool, vips := range desired.Spec {
			enrichVIPsCache[deployPool] = addToWhitelists(true, vips, addedCIDRs...)
		}
		return enrichVIPsCache, authCheckResultsCache, nil
	}
}
