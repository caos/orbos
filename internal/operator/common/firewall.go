package common

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Firewall struct {
	mux   sync.Mutex       `yaml:"-"`
	Zones map[string]*Zone `yaml:",inline"`
}

type Service struct {
	Description string
	Ports       []*Allowed
}

func ToFirewall(zone string, fw map[string]*Allowed) Firewall {
	return Firewall{Zones: map[string]*Zone{zone: {Interfaces: []string{}, FW: fw, Services: map[string]*Service{}}}}
}

func (f *Firewall) Merge(fw Firewall) {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.Zones == nil {
		f.Zones = make(map[string]*Zone, 0)
	}

	if fw.Zones == nil {
		return
	}

	for name, zone := range fw.Zones {
		if zone == nil {
			continue
		}

		current, ok := f.Zones[name]
		if !ok || current == nil {
			current = &Zone{}
		}

		if current.Masquerade || zone.Masquerade {
			current.Masquerade = true
		}

		if zone.FW != nil {
			if current.FW == nil {
				current.FW = make(map[string]*Allowed)
			}

			for key, value := range zone.FW {
				current.FW[key] = value
			}
		}

		if zone.Interfaces != nil {
			if current.Interfaces == nil {
				current.Interfaces = []string{}
			}
			for _, i := range zone.Interfaces {
				found := false
				for _, i2 := range current.Interfaces {
					if i == i2 {
						found = true
					}
				}
				if !found {
					current.Interfaces = append(current.Interfaces, i)
				}
			}

		}
		if zone.Sources != nil {
			if current.Sources == nil {
				current.Sources = make([]string, 0)
			}
			for _, i := range zone.Sources {
				found := false
				for _, i2 := range current.Sources {
					if i == i2 {
						found = true
					}
				}
				if !found {
					current.Sources = append(current.Sources, i)
				}
			}
		}
		if zone.Services != nil {
			if current.Services == nil {
				current.Services = make(map[string]*Service, 0)
			}

			for key, value := range zone.Services {
				current.Services[key] = value
			}
		}

		f.Zones[name] = current
	}
}

func (f *Firewall) ToCurrent() FirewallCurrent {
	zones := make(FirewallCurrent, 0)
	if f.Zones == nil {
		return zones
	}

	for name, zone := range f.Zones {
		if zone != nil {
			zones = append(zones, &ZoneDesc{
				Name:       name,
				Masquerade: zone.Masquerade,
				Interfaces: zone.Interfaces,
				Sources:    zone.Sources,
				Services:   []*Service{},
				FW:         f.Ports(name),
			})
		}
	}
	zones.Sort()
	return zones
}

func (f *Firewall) Ports(zoneName string) Ports {
	ports := make([]*Allowed, 0)
	if f.Zones == nil {
		return ports
	}
	for name, zone := range f.Zones {
		if name == zoneName && zone != nil && zone.FW != nil {
			for _, value := range zone.FW {
				ports = append(ports, value)
			}
		}
	}
	return ports
}

type ZoneDesc struct {
	Name       string
	Masquerade bool
	Interfaces MarshallableSlice
	Sources    MarshallableSlice
	FW         []*Allowed
	Services   []*Service
}

type Zone struct {
	Masquerade bool
	Interfaces MarshallableSlice
	Sources    MarshallableSlice
	FW         map[string]*Allowed
	Services   map[string]*Service
}

type FirewallCurrent []*ZoneDesc

var _ fmt.Stringer = (*FirewallCurrent)(nil)

func (c FirewallCurrent) String() string {
	fw := ""
	for _, zone := range c {

		masquerade := ""
		if zone.Masquerade {
			masquerade = " masquerade: true "
		}

		interfaces := ""
		for idx := range zone.Interfaces {
			interfaces = interfaces + zone.Interfaces[idx] + " "
		}

		sources := ""
		for idx := range zone.Sources {
			sources = sources + zone.Sources[idx] + " "
		}

		ports := ""
		for _, port := range zone.FW {
			ports = ports + port.Port + "/" + port.Protocol + " "
		}
		fw = fw + zone.Name + masquerade + "(" + strings.TrimSpace(strings.TrimSpace(strings.TrimSpace(interfaces)+" "+sources)+" "+ports) + ") "
	}
	return strings.TrimSpace(fw)
}

func (c FirewallCurrent) Sort() {
	sort.Slice(c, func(i, j int) bool {
		return c[i].Name < c[j].Name
	})

	for _, currentEntry := range c {

		sort.Slice(currentEntry.Interfaces, func(i, j int) bool {
			return currentEntry.Interfaces[i] < currentEntry.Interfaces[j]
		})

		sort.Slice(currentEntry.FW, func(i, j int) bool {
			iEntry := currentEntry.FW[i].Port + "/" + currentEntry.FW[i].Protocol
			jEntry := currentEntry.FW[j].Port + "/" + currentEntry.FW[j].Protocol
			return iEntry < jEntry
		})

		sort.Slice(currentEntry.Services, func(i, j int) bool {
			return currentEntry.Services[i].Description < currentEntry.Services[j].Description
		})

		for _, svc := range currentEntry.Services {
			sort.Slice(svc.Ports, func(i, j int) bool {
				iEntry := svc.Ports[i].Port + "/" + svc.Ports[i].Protocol
				jEntry := svc.Ports[j].Port + "/" + svc.Ports[j].Protocol
				return iEntry < jEntry
			})
		}

		sort.Slice(currentEntry.Sources, func(i, j int) bool {
			return currentEntry.Sources[i] < currentEntry.Sources[j]
		})
	}
}

type Ports []*Allowed

func (p Ports) String() string {
	strs := make([]string, len(p))
	for idx, port := range p {
		strs[idx] = fmt.Sprintf("%s/%s", port.Port, port.Protocol)
	}
	return strings.Join(strs, " ")
}

type Allowed struct {
	Port     string
	Protocol string
}

func (f Firewall) Contains(other Firewall) bool {
	if other.Zones == nil {
		return true
	}

	for name, zone := range other.Zones {
		current, ok := f.Zones[name]
		if !ok || current == nil {
			return false
		}

		if zone.Masquerade != current.Masquerade {
			return false
		}

		if zone.FW != nil && len(zone.FW) > 0 {
			if current.FW == nil || len(current.FW) == 0 {
				return false
			}

			for name, port := range zone.FW {
				found, ok := current.FW[name]
				if !ok {
					return false
				}
				if !deriveEqualPort(*port, *found) {
					return false
				}
			}
		}

		if zone.Sources != nil && len(zone.Sources) > 0 {
			if current.Sources == nil || len(current.Sources) == 0 {
				return false
			}
			for _, source := range zone.Sources {
				found := false
				for _, currentSource := range current.Sources {
					if currentSource == source {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}
	return true
}

func (f Firewall) IsContainedIn(zones FirewallCurrent) bool {
	if f.Zones == nil || len(f.Zones) == 0 {
		return true
	}
	if zones == nil || len(zones) == 0 {
		return false
	}

	for name, zone := range f.Zones {
		if (zone.FW == nil || len(zone.FW) == 0) &&
			(zone.Sources == nil || len(zone.Sources) == 0) {
			continue
		}

		foundZone := false
		for _, currentZone := range zones {
			if currentZone == nil {
				continue
			}

			if foundZone {
				break
			}

			if currentZone.Name == name {
				foundZone = true

				if zone.Masquerade && !currentZone.Masquerade {
					return false
				}

				if currentZone.FW == nil || len(currentZone.FW) == 0 {
					return false
				}

				if zone.FW != nil {
					for _, fwPort := range zone.FW {
						foundPort := false
						if currentZone.FW != nil {
							for _, currentPort := range currentZone.FW {
								if deriveEqualPort(*currentPort, *fwPort) {
									foundPort = true
									break
								}
							}
						}

						if !foundPort {
							return false
						}
					}
				}

				if zone.Sources != nil {
					for _, source := range zone.Sources {
						foundSource := false
						if currentZone.Sources != nil {
							for _, currentSource := range currentZone.Sources {
								if source == currentSource {
									foundSource = true
									break
								}
							}
						}
						if !foundSource {
							return false
						}
					}
				}
			}
		}
		if !foundZone {
			return false
		}
	}
	return true
}
