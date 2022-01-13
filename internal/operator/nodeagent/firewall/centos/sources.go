package centos

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
)

func getAddAndRemoveSources(
	monitor mntr.Monitor,
	zoneName string,
	current *common.ZoneDesc,
	desired common.Firewall,
) (
	[]string,
	[]string,
) {

	addSources := make([]string, 0)
	removeSources := make([]string, 0)
	zone := desired.Zones[zoneName]

	if zone.Sources != nil && len(zone.Sources) > 0 {
		for _, source := range zone.Sources {
			foundSource := false
			if current.Sources != nil && len(current.Sources) > 0 {
				for _, currentSource := range current.Sources {
					if currentSource == source {
						foundSource = true
					}
				}
			}
			if !foundSource {
				addSources = append(addSources, fmt.Sprintf("--add-source=%s", source))
			}
		}
	}
	if current.Sources != nil && len(current.Sources) > 0 {
		for _, currentSource := range current.Sources {
			foundSource := false
			if zone.Sources != nil && len(zone.Sources) > 0 {
				for _, source := range zone.Sources {
					if source == currentSource {
						foundSource = true
					}
				}
			}
			if !foundSource {
				removeSources = append(removeSources, fmt.Sprintf("--remove-source=%s", currentSource))
			}
		}
	}

	return addSources, removeSources
}
