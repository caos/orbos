package dynamic

import (
	"bytes"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/mntr"
)

var	probes = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name:       "probe",
		Help:       "Load Balancing Probes.",
	},
	[]string{"target"},
)

func init(){
	prometheus.MustRegister(probes)
}

func AdaptFunc() orbiter.AdaptFunc {
	return func(monitor mntr.Monitor, desiredTree *orbiter.Tree, currentTree *orbiter.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, secrets map[string]*orbiter.Secret, migrate bool, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, nil, migrate, errors.Wrapf(err, "unmarshaling desired state for kind %s failed", desiredTree.Common.Kind)
		}
		if err := desiredKind.Validate(); err != nil {
			return nil, nil, nil, migrate, err
		}
		desiredKind.Common.Version = "v0"
		desiredTree.Parsed = desiredKind

		current := &Current{
			Common: &orbiter.Common{
				Kind:    "orbiter.caos.ch/DynamicLoadBalancer",
				Version: "v0",
			},
		}
		currentTree.Parsed = current

		return func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec, queried map[string]interface{}) (orbiter.EnsureFunc, error) {

			sourcePools := make(map[string][]string)
			addresses := make(map[string]infra.Address)
			for _, pool := range desiredKind.Spec {
				for _, vip := range pool {
					for _, src := range vip.Transport {
						addresses[src.Name] = infra.Address{
							Location: vip.IP,
							Port:     uint16(src.SourcePort),
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
			current.Current.Addresses = addresses
			current.Current.Desire = func(pool string, svc core.MachinesService, nodeagents map[string]*common.NodeAgentSpec, notifyMaster string) error {

				vips, ok := desiredKind.Spec[pool]
				if !ok {
					return nil
				}

				machines, err := svc.List(pool, true)
				if err != nil {
					return err
				}

				machinesData := make([]Data, len(machines))
				for idx, machine := range machines {
					machinesData[idx] = Data{
						VIPs: vips,
						Self: machine,
						Peers: deriveFilterMachines(func(cmp infra.Machine) bool {
							return cmp.ID() != machine.ID()
						}, append([]infra.Machine(nil), []infra.Machine(machines)...)),
						State:                "BACKUP",
						CustomMasterNotifyer: notifyMaster != "",
					}
					if idx == 0 {
						machinesData[idx].State = "MASTER"
					}
				}

				templateFuncs := template.FuncMap(map[string]interface{}{
					"machines": svc.List,
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

{{ range $idx, $vip := .VIPs }}vrrp_script chk_{{ $vip.IP }} {
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
		chk_{{ $vip.IP }}
	}

{{ if $root.CustomMasterNotifyer }}	notify_master "/etc/keepalived/notifymaster.sh {{ $root.Self.ID }} {{ $vip.IP }}"
{{ else }}	virtual_ipaddress {
		{{ $vip.IP }}
	}
{{ end }}
}
{{ end }}
`))

				nginxTemplate := template.Must(template.New("").Funcs(templateFuncs).Parse(`{{ $root := . }}events {
	worker_connections  4096;  ## Default: 1024
}

stream { {{ range $vip := .VIPs }}{{ range $src := $vip.Transport }}
	upstream {{ $src.Name }} {    {{ range $dest := $src.Destinations }}{{ range $machine := machines $dest.Pool true }}
		server {{ $machine.IP }}:{{ if eq $src.Name  "kubeapi" }}6666{{ else }}{{ $dest.Port }}{{ end }}; # {{ $dest.Pool }}{{end}}{{ end }}
	}
	server {
		listen {{ $vip.IP }}:{{ $src.SourcePort }};
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

				for _, d := range machinesData {

					var kaBuf bytes.Buffer
					if err := keepaliveDTemplate.Execute(&kaBuf, d); err != nil {
						return err
					}
					kaPkg := common.Package{Config: map[string]string{"keepalived.conf": kaBuf.String()}}

					if d.CustomMasterNotifyer {
						kaPkg.Config["notifymaster.sh"] = notifyMaster
					}

					for _, vip := range d.VIPs {
						for _, transport := range vip.Transport {
							for _, machine := range machines {
								deepNa, ok := nodeagents[machine.ID()]
								if !ok {
									deepNa = &common.NodeAgentSpec{}
									nodeagents[machine.ID()] = deepNa
								}
								if deepNa.Firewall == nil {
									deepNa.Firewall = &common.Firewall{}
								}
								fw := *deepNa.Firewall
								fw[fmt.Sprintf("%s-%d-src", transport.Name, transport.SourcePort)] = common.Allowed{
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
					na, ok := nodeagents[d.Self.ID()]
					if !ok {
						na = &common.NodeAgentSpec{}
						nodeagents[d.Self.ID()] = na
					}
					if na.Software == nil {
						na.Software = &common.Software{}
					}
					na.Software.KeepaliveD = kaPkg

					var ngxBuf bytes.Buffer
					if nginxTemplate.Execute(&ngxBuf, d); err != nil {
						return err
					}
					ngxPkg := common.Package{Config: map[string]string{"nginx.conf": ngxBuf.String()}}
					for _, vip := range d.VIPs {
						for _, transport := range vip.Transport {
							vipProbe := fmt.Sprintf("%s:%d", vip.IP, transport.SourcePort)
							probes.With(prometheus.Labels{
								"target": fmt.Sprintf("%s VIP (%s)",transport.Name, vipProbe),
							})
							transport.Destinations[0].HealthChecks.

							for _, dest := range transport.Destinations {
								destMachines, err := svc.List(dest.Pool, true)
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
									fw[fmt.Sprintf("%s-%d-dest", transport.Name, dest.Port)] = common.Allowed{
										Port:     fmt.Sprintf("%d", dest.Port),
										Protocol: "tcp",
									}
								}
							}
						}
					}
					na.Software.Nginx = ngxPkg
				}
				return nil
			}
			return nil, nil
		}, nil, nil, false, nil
	}
}

type Data struct {
	VIPs                 []*VIP
	State                string
	RouterID             int
	Self                 infra.Machine
	Peers                []infra.Machine
	CustomMasterNotifyer bool
}
