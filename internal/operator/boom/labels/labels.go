package labels

import "github.com/caos/orbos/internal/operator/boom/name"

var (
	instanceName = "boom"
)

func GetGlobalLabels() map[string]string {
	labels := make(map[string]string, 0)
	labels["app.kubernetes.io/managed-by"] = "boom.caos.ch"
	labels["boom.caos.ch/part-of"] = "boom"
	labels["boom.caos.ch/instance"] = instanceName

	return labels
}

func GetAllApplicationLabels(appName name.Application) map[string]string {
	labels := GetGlobalLabels()
	addLabels := GetApplicationLabels(appName)

	for k, v := range addLabels {
		labels[k] = v
	}
	return labels
}

func GetApplicationLabels(appName name.Application) map[string]string {
	labels := make(map[string]string, 0)
	labels["boom.caos.ch/application"] = appName.String()
	return labels
}

func GetAllForApplicationLabels(appName name.Application) map[string]string {
	labels := GetGlobalLabels()
	addLabels := GetForApplicationLabels(appName)

	for k, v := range addLabels {
		labels[k] = v
	}
	return labels
}

func GetForApplicationLabels(appName name.Application) map[string]string {
	labels := make(map[string]string, 0)
	labels["boom.caos.ch/for-application"] = appName.String()
	return labels
}

func GetMonitorLabels(instanceName string, appName name.Application) map[string]string {
	labels := GetApplicationLabels(appName)
	addLabels := GetMonitorSelectorLabels(instanceName)

	for k, v := range addLabels {
		labels[k] = v
	}
	return labels
}

func GetMonitorSelectorLabels(instanceName string) map[string]string {
	labels := make(map[string]string, 0)
	labels["boom.caos.ch/prometheus"] = instanceName
	return labels
}

func GetRuleLabels(instanceName string, appName name.Application) map[string]string {
	labels := GetApplicationLabels(appName)
	addLabels := GetRuleSelectorLabels(instanceName)

	for k, v := range addLabels {
		labels[k] = v
	}
	return labels
}

func GetRuleSelectorLabels(instanceName string) map[string]string {
	labels := make(map[string]string, 0)
	labels["boom.caos.ch/prometheus"] = instanceName
	return labels
}
