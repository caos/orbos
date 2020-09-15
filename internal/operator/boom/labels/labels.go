package labels

import "github.com/caos/orbos/internal/operator/boom/name"

var (
	instanceName = "boom"
)

func GetGlobalLabels() Labels {
	return map[string]string{
		"app.kubernetes.io/managed-by": "boom.caos.ch",
		"boom.caos.ch/part-of":         "boom",
		"boom.caos.ch/instance":        instanceName,
	}
}

func GetAllApplicationLabels(appName name.Application) Labels {
	return GetGlobalLabels().
		Append(GetApplicationLabels(appName))
}

func GetApplicationLabels(appName name.Application) Labels {
	return map[string]string{
		"boom.caos.ch/application": appName.String(),
	}
}

func GetPromSelector(instanceName string) Labels {
	return map[string]string{
		"boom.caos.ch/prometheus": instanceName,
	}
}

func GetMonitorLabels(instanceName string, appName name.Application) map[string]string {
	return GetApplicationLabels(appName).
		Append(GetPromSelector(instanceName))
}

type Labels map[string]string

func (l Labels) Append(labels map[string]string) Labels {
	for k, v := range labels {
		l[k] = v
	}
	return l
}
