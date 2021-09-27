package current

import (
	"sort"

	"github.com/caos/orbos/v5/internal/operator/boom/labels"
	"github.com/caos/orbos/v5/internal/operator/boom/name"
	"github.com/caos/orbos/v5/internal/utils/clientgo"
	"github.com/caos/orbos/v5/mntr"
)

type Current struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Resources  []*clientgo.Resource
}

func ResourcesToYaml(resources []*clientgo.Resource) *Current {
	return &Current{
		APIVersion: "boom.caos.ch/v1beta1",
		Kind:       "currentstate",
		Resources:  resources,
	}
}

func Get(monitor mntr.Monitor, resourceInfoList []*clientgo.ResourceInfo) []*clientgo.Resource {
	globalLabels := labels.GetGlobalLabels()

	resources, err := clientgo.ListResources(monitor, resourceInfoList, globalLabels)
	if err != nil {
		return nil
	}

	sort.Sort(clientgo.ResourceSorter(resources))

	return resources
}

func FilterForApplication(appName name.Application, currentResourceList []*clientgo.Resource) []*clientgo.Resource {
	filteredResourceList := make([]*clientgo.Resource, 0)

	for _, currentResource := range currentResourceList {
		applicationLabels := labels.GetAllApplicationLabels(appName)
		// forApplicationLabels := labels.GetAllForApplicationLabels(appName)

		//determine if currentresources are with application or for-application-labels
		allfound := true
		// application
		for applicationLabel, applicationLabelValue := range applicationLabels {
			found := false
			for currentResourceLabel, currentResourceLabelValue := range currentResource.Labels {
				if applicationLabel == currentResourceLabel &&
					applicationLabelValue == currentResourceLabelValue {
					found = true
					break
				}
			}
			if found == false {
				allfound = false
				break
			}
		}
		// if found not necessary to check for-application-labels
		if allfound == true {
			filteredResourceList = append(filteredResourceList, currentResource)
			continue
		}
		// // for-application
		// for forApplicationLabel, forApplicationLabelValue := range forApplicationLabels {
		// 	for currentResourceLabel, currentResourceLabelValue := range currentResource.Labels {
		// 		if forApplicationLabel == currentResourceLabel &&
		// 			forApplicationLabelValue == currentResourceLabelValue {
		// 			found = true
		// 			break
		// 		}
		// 	}
		// 	if found == false {
		// 		break
		// 	}
		// }
		// if found == true {
		// 	filteredResourceList = append(filteredResourceList, currentResource)
		// }
	}

	return filteredResourceList
}
