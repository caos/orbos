package centos

import (
	"strings"

	"github.com/caos/orbos/mntr"
	"gopkg.in/yaml.v3"
)

type commaSeparatedStrings struct {
	slice []string
}

func (c *commaSeparatedStrings) UnmarshalYAML(node *yaml.Node) error {

	var str string

	if err := node.Decode(&str); err != nil {
		return err
	}

	c.slice = strings.Fields(str)
	return nil
}

type Zone struct {
	Target     string
	Interfaces commaSeparatedStrings
	Sources    commaSeparatedStrings
	Ports      commaSeparatedStrings
	Protocols  commaSeparatedStrings
	Masquerade bool
}

func queryCurrentFirewall(monitor mntr.Monitor) (map[string]Zone, error) {

	allZones, err := runFirewallCommand(monitor, "--list-all-zones")
	if err != nil {
		return nil, err
	}

	zoneStrings := strings.Split(allZones, "\t\n\n")

	zones := make(map[string]Zone)
	for _, zoneString := range zoneStrings {
		firstLineIdx := strings.Index(zoneString, "\n")
		zoneName := strings.Fields(zoneString[:firstLineIdx])[0]
		zone := Zone{}

		prunedZone := strings.ReplaceAll(zoneString[firstLineIdx:], "\t", "")
		prunedZone = strings.ReplaceAll(prunedZone, "%%REJECT%%", `"%%REJECT%%"`)
		if err := yaml.Unmarshal([]byte(prunedZone), &zone); err != nil {
			panic(err)
		}
		zones[zoneName] = zone
	}
	return zones, err
}
