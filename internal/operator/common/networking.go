package common

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Networking struct {
	mux        sync.Mutex `yaml:"-"`
	Interfaces map[string]*NetworkingInterface
}

type NetworkingInterface struct {
	Type string
	IPs  MarshallableSlice
}

type NetworkingCurrent []*NetworkingInterfaceCurrent

type NetworkingInterfaceCurrent struct {
	Name string
	IPs  MarshallableSlice
}

func (n *Networking) Merge(nw Networking) {
	n.mux.Lock()
	defer n.mux.Unlock()
	if n.Interfaces == nil {
		n.Interfaces = make(map[string]*NetworkingInterface, 0)
	}

	if nw.Interfaces == nil {
		return
	}

	for name, iface := range nw.Interfaces {
		if iface == nil {
			continue
		}
		current, ok := n.Interfaces[name]
		if !ok || current == nil {
			current = &NetworkingInterface{}
		}

		current.Type = iface.Type

		if iface.IPs != nil {
			if current.IPs == nil {
				current.IPs = make(MarshallableSlice, 0)
			}

			for _, value := range iface.IPs {
				found := false
				for _, currentValue := range current.IPs {
					if currentValue == value {
						found = true
					}
				}
				if !found {
					current.IPs = append(current.IPs, value)
				}
			}
		}
		n.Interfaces[name] = current
	}
}

func (n *Networking) ToCurrent() NetworkingCurrent {
	current := make(NetworkingCurrent, 0)
	if n.Interfaces == nil {
		return current
	}

	for name, iface := range n.Interfaces {
		if iface != nil {
			current = append(current, &NetworkingInterfaceCurrent{
				Name: name,
				IPs:  iface.IPs,
			})
		}
	}
	current.Sort()
	return current
}

func (c NetworkingCurrent) Sort() {
	sort.Slice(c, func(i, j int) bool {
		return c[i].Name < c[j].Name
	})

	for _, currentEntry := range c {
		sort.Slice(currentEntry.IPs, func(i, j int) bool {
			return currentEntry.IPs[i] < currentEntry.IPs[j]
		})
	}
}

func (n Networking) IsContainedIn(interfaces NetworkingCurrent) bool {
	if n.Interfaces == nil || len(n.Interfaces) == 0 {
		return true
	}
	if interfaces == nil || len(interfaces) == 0 {
		return false
	}

	for name, iface := range n.Interfaces {
		if iface.IPs == nil || len(iface.IPs) == 0 {
			continue
		}

		foundIface := false
		for _, currentInterface := range interfaces {
			if currentInterface == nil {
				continue
			}

			if foundIface {
				break
			}

			if currentInterface.Name == name {
				foundIface = true

				if currentInterface.IPs == nil || len(currentInterface.IPs) == 0 {
					return false
				}

				if iface.IPs != nil {
					for _, ip := range iface.IPs {
						foundIP := false
						if currentInterface.IPs != nil {
							for _, currentIP := range currentInterface.IPs {
								if ip == currentIP {
									foundIP = true
									break
								}
							}
						}

						if !foundIP {
							return false
						}
					}
				}
			}
		}
		if !foundIface {
			return false
		}
	}
	return true
}

var _ fmt.Stringer = (*NetworkingCurrent)(nil)

func (c NetworkingCurrent) String() string {
	nw := ""
	for _, iface := range c {

		ips := ""
		for idx := range iface.IPs {
			ips = ips + iface.IPs[idx] + " "
		}
		nw = nw + iface.Name + "(" + strings.TrimSpace(ips) + ") "
	}
	return strings.TrimSpace(nw)
}
