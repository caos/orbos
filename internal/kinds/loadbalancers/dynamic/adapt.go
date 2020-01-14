package dynamic

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/core"
)

func AdaptFunc(remoteUser string) orbiter.AdaptFunc {
	return func(desiredTree *orbiter.Tree, secretsTree *orbiter.Tree, currentTree *orbiter.Tree) (ensureFunc orbiter.EnsureFunc, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()
		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, errors.Wrapf(err, "unmarshaling desired state for kind %s failed", desiredTree.Common.Kind)
		}
		if err := desiredKind.Validate(); err != nil {
			return nil, err
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
		current.Current.Desire = func(pool string, svc core.ComputesService, nodeagents map[string]*orbiter.NodeAgentSpec, notifyMaster string) error {

			vips, ok := desiredKind.Spec[pool]
			if !ok {
				return nil
			}

			computes, err := svc.List(pool, true)
			if err != nil {
				return err
			}

			computesData := make([]Data, len(computes))
			for idx, compute := range computes {
				computesData[idx] = Data{
					RemoteUser: remoteUser,
					VIPs:       vips,
					Self:       compute,
					Peers: deriveFilterComputes(func(cmp infra.Compute) bool {
						return cmp.ID() != compute.ID()
					}, append([]infra.Compute(nil), []infra.Compute(computes)...)),
					State:                "BACKUP",
					CustomMasterNotifyer: notifyMaster != "",
				}
				if idx == 0 {
					computesData[idx].State = "MASTER"
				}
			}

			templateFuncs := template.FuncMap(map[string]interface{}{
				"computes": svc.List,
				"add": func(i, y int) int {
					return i + y
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
	script_user {{ $root.RemoteUser }}
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
	upstream {{ $src.Name }} {    {{ range $dest := $src.Destinations }}{{ range $compute := computes $dest.Pool true }}
		server {{ $compute.IP }}:{{ if eq $src.Name  "kubeapi" }}6666{{ else }}{{ $dest.Port }}{{ end }}; # {{ $dest.Pool }}{{end}}{{ end }}
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

			for _, d := range computesData {

				var kaBuf bytes.Buffer
				if err := keepaliveDTemplate.Execute(&kaBuf, d); err != nil {
					return err
				}
				kaPkg := orbiter.Package{Config: map[string]string{"keepalived.conf": kaBuf.String()}}

				if d.CustomMasterNotifyer {
					kaPkg.Config["notifymaster.sh"] = notifyMaster
				}

				for _, vip := range d.VIPs {
					for _, transport := range vip.Transport {
						for _, compute := range computes {
							deepNa, ok := nodeagents[compute.ID()]
							if !ok {
								deepNa = &orbiter.NodeAgentSpec{}
								nodeagents[compute.ID()] = deepNa
							}

							deepNa.Firewall[fmt.Sprintf("%s-%d-src", transport.Name, transport.SourcePort)] = orbiter.Allowed{
								Port:     fmt.Sprintf("%d", transport.SourcePort),
								Protocol: "tcp",
							}
						}
					}
				}
				na, ok := nodeagents[d.Self.ID()]
				if !ok {
					na = &orbiter.NodeAgentSpec{}
					nodeagents[d.Self.ID()] = na
				}
				if na.Software == nil {
					na.Software = &orbiter.Software{}
				}
				na.Software.KeepaliveD = kaPkg

				var ngxBuf bytes.Buffer
				if nginxTemplate.Execute(&ngxBuf, d); err != nil {
					return err
				}
				ngxPkg := orbiter.Package{Config: map[string]string{"nginx.conf": ngxBuf.String()}}
				for _, vip := range d.VIPs {
					for _, transport := range vip.Transport {
						for _, dest := range transport.Destinations {
							destComputes, err := svc.List(dest.Pool, true)
							if err != nil {
								return err
							}
							for _, compute := range destComputes {

								deepNa, ok := nodeagents[compute.ID()]
								if !ok {
									deepNa = &orbiter.NodeAgentSpec{}
									nodeagents[compute.ID()] = deepNa
								}

								deepNa.Firewall[fmt.Sprintf("%s-%d-dest", transport.Name, dest.Port)] = orbiter.Allowed{
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
	}
}

type Data struct {
	VIPs                 []VIP
	RemoteUser           string
	State                string
	RouterID             int
	Self                 infra.Compute
	Peers                []infra.Compute
	CustomMasterNotifyer bool
}
