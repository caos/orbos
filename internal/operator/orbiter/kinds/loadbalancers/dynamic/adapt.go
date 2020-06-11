package dynamic

import (
	"bytes"
	"fmt"
	"sort"
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

	return func(monitor mntr.Monitor, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, migrate bool, err error) {

		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		if desiredTree.Common.Version != "v1" {
			migrate = true
		}
		desiredKind := &Desired{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, migrate, errors.Wrapf(err, "unmarshaling desired state for kind %s failed", desiredTree.Common.Kind)
		}

		for _, pool := range desiredKind.Spec {
			for _, vip := range pool {
				for _, src := range vip.Transport {
					for _, dest := range src.Destinations {
						if src.Name == "kubeapi" {
							if dest.Port != 6666 {
								dest.Port = 6666
								migrate = true
							}
							if dest.HealthChecks.Path != "/healthz" {
								dest.HealthChecks.Path = "/healthz"
								migrate = true
							}
							if dest.HealthChecks.Protocol != "https" {
								dest.HealthChecks.Protocol = "https"
								migrate = true
							}
						}
					}
					if len(src.Whitelist) == 0 {
						allIPs := orbiter.CIDR("0.0.0.0/0")
						src.Whitelist = []*orbiter.CIDR{&allIPs}
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

		return func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec, queried map[string]interface{}) (orbiter.EnsureFunc, error) {

			wl := whitelist()

			sourcePools := make(map[string][]string)
			addresses := make(map[string]*infra.Address)
			for _, pool := range desiredKind.Spec {
				for _, vip := range pool {
					for _, src := range vip.Transport {
						addresses[src.Name] = &infra.Address{
							Location:     vip.IP,
							ExternalPort: uint16(src.SourcePort),
						}
					destinations:
						for _, dest := range src.Destinations {
							if _, ok := sourcePools[dest.Pool]; !ok {
								sourcePools[dest.Pool] = make([]string, 0)
							}
							for _, existing := range sourcePools[dest.Pool] {
								if dest.Pool == existing {
									continue destinations
								}
							}
							sourcePools[dest.Pool] = append(sourcePools[dest.Pool], dest.Pool)
						}
					}
				}
			}

			current.Current.SourcePools = sourcePools
			current.Current.Spec = desiredKind.Spec
			current.Current.Desire = func(forPool string, svc core.MachinesService, nodeagents map[string]*common.NodeAgentSpec, hyperconverged bool, notifyMaster func(machine infra.Machine, peers infra.Machines, vips []*VIP) string, mapVIP func(*VIP) string) error {

				vips, ok := desiredKind.Spec[forPool]
				if !ok {
					return nil
				}

				allPools, err := svc.ListPools()
				if err != nil {
					return err
				}

				var forMachines infra.Machines
				for _, pool := range allPools {
					machines, err := svc.List(pool)
					if err != nil {
						return err
					}
					if forPool == pool {
						forMachines = machines
					}
					for _, machine := range machines {
						cidr := orbiter.CIDR(fmt.Sprintf("%s/32", machine.IP()))
						vips = addToWhitelists(false, vips, &cidr)
					}
				}

				vips = addToWhitelists(true, vips, wl...)

				machinesData := make([]Data, len(forMachines))
				for idx, machine := range forMachines {
					machinesData[idx] = Data{
						VIPs: vips,
						Self: machine,
						Peers: deriveFilterMachines(func(cmp infra.Machine) bool {
							return cmp.ID() != machine.ID()
						}, append([]infra.Machine(nil), []infra.Machine(forMachines)...)),
						State:                "BACKUP",
						CustomMasterNotifyer: notifyMaster != nil,
					}
					if idx == 0 {
						machinesData[idx].State = "MASTER"
					}

					for _, ignoredHyperConvergedDeployPool := range desiredKind.Spec {
						for _, vip := range ignoredHyperConvergedDeployPool {
							for _, src := range vip.Transport {
								for _, dest := range src.Destinations {
									if dest.Pool == forPool {
										machinesData[idx].NATs = append(machinesData[idx].NATs, &NAT{
											Name: src.Name,
											From: []string{
												fmt.Sprintf("%s:%d", mapVIP(vip), src.SourcePort),  // VIP
												fmt.Sprintf("%s:%d", machine.IP(), src.SourcePort), // Node IP
											},
											To: fmt.Sprintf("127.0.0.1:%d", dest.Port),
										})
										break
									}
								}
							}
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
					//						"healthcmd": vrrpHealthChecksScript,
					//						"upstreamHealthchecks": deriveFmap(vip model.VIP) []string {
					//							return deriveFmap(func(src model.Source) []string {
					//
					//								if src.HealthChecks != nil {
					//									return fmt.Sprintf(check, src.HealthChecks.Protocol)
					//								}
					//							}, vip.Transport)
					//						},
				})

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
	fall 15      # require 2 failures for KO
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
	upstream {{ $src.Name }} {    {{ range $dest := $src.Destinations }}{{ range $machine := forMachines $dest.Pool }}
		server {{ $machine.IP }}:{{ $dest.Port }}; # {{ $dest.Pool }}{{end}}{{ end }}
	}
	server {
		listen {{ vip $vip }}:{{ $src.SourcePort }};
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

				nginxNATTemplate := template.Must(template.New("").Funcs(templateFuncs).Parse(`{{ $root := . }}events {
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

				for _, d := range machinesData {

					na, ok := nodeagents[d.Self.ID()]
					if !ok {
						na = &common.NodeAgentSpec{}
						nodeagents[d.Self.ID()] = na
					}
					if na.Software == nil {
						na.Software = &common.Software{}
					}

					ngxBuf := new(bytes.Buffer)
					defer ngxBuf.Reset()

					if !hyperconverged {
						if err := nginxNATTemplate.Execute(ngxBuf, d); err != nil {
							return err
						}
					} else {
						kaBuf := new(bytes.Buffer)
						defer kaBuf.Reset()

						if err := keepaliveDTemplate.Execute(kaBuf, d); err != nil {
							return err
						}

						for _, vip := range d.VIPs {
							if vip.IP == "" {
								return errors.New("No IP configured")
							}
							for _, transport := range vip.Transport {
								if err := transport.SourcePort.validate(); err != nil {
									return err
								}
								for _, machine := range forMachines {
									deepNa, ok := nodeagents[machine.ID()]
									if !ok {
										deepNa = &common.NodeAgentSpec{}
										nodeagents[machine.ID()] = deepNa
									}
									if deepNa.Firewall == nil {
										deepNa.Firewall = &common.Firewall{}
									}
									fw := *deepNa.Firewall
									fw[fmt.Sprintf("%s-%d-src", transport.Name, transport.SourcePort)] = &common.Allowed{
										Port:     fmt.Sprintf("%d", transport.SourcePort),
										Protocol: "tcp",
									}

									if transport.SourcePort == 22 {
										if deepNa.Software == nil {
											deepNa.Software = &common.Software{}
										}
										deepNa.Software.SSHD.Config = map[string]string{"listenaddress": machine.IP()}
									}
								}
							}
						}

						kaPkg := common.Package{Config: map[string]string{"keepalived.conf": kaBuf.String()}}

						if d.CustomMasterNotifyer {
							kaPkg.Config["notifymaster.sh"] = notifyMaster(d.Self, d.Peers, d.VIPs)
						}

						na.Software.KeepaliveD = kaPkg

						if err := nginxLBTemplate.Execute(ngxBuf, d); err != nil {
							return err
						}
						for _, vip := range d.VIPs {
							for _, transport := range vip.Transport {
								probe("VIP", vip.IP, uint16(transport.SourcePort), transport.Destinations[0].HealthChecks, *transport)
								for _, dest := range transport.Destinations {
									destMachines, err := svc.List(dest.Pool)
									if err != nil {
										return err
									}
									for _, machine := range destMachines {

										deepNa, ok := nodeagents[machine.ID()]
										if !ok {
											deepNa = &common.NodeAgentSpec{}
											nodeagents[machine.ID()] = deepNa
										}
										if deepNa.Firewall == nil {
											deepNa.Firewall = &common.Firewall{}
										}
										fw := *deepNa.Firewall
										fw[fmt.Sprintf("%s-%d-dest", transport.Name, dest.Port)] = &common.Allowed{
											Port:     fmt.Sprintf("%d", dest.Port),
											Protocol: "tcp",
										}
										probe("Upstream", machine.IP(), uint16(dest.Port), dest.HealthChecks, *transport)
									}
								}
							}
						}
					}
					ngxPkg := common.Package{Config: map[string]string{"nginx.conf": ngxBuf.String()}}
					na.Software.Nginx = ngxPkg
					sysctl.SetProperty(&na.Software.Sysctl, sysctl.IpForward, true)
					sysctl.SetProperty(&na.Software.Sysctl, sysctl.NonLocalBind, true)
				}
				return nil
			}
			return nil, nil
		}, nil, migrate, nil
	}
}

func addToWhitelists(makeUnique bool, vips []*VIP, cidr ...*orbiter.CIDR) []*VIP {
	newVIPs := make([]*VIP, len(vips))
	for vipIdx, vip := range vips {
		newTransport := make([]*Source, len(vip.Transport))
		for srcIdx, src := range vip.Transport {
			newSource := &Source{
				Name:         src.Name,
				SourcePort:   src.SourcePort,
				Destinations: src.Destinations,
				Whitelist:    append(src.Whitelist, cidr...),
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

func probe(probeType, ip string, port uint16, hc HealthChecks, source Source) {
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

type NAT struct {
	Name string
	From []string
	To   string
}

type Data struct {
	VIPs                 []*VIP
	State                string
	RouterID             int
	Self                 infra.Machine
	Peers                []infra.Machine
	CustomMasterNotifyer bool
	NATs                 []*NAT
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
