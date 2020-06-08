package metrics

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetServicemonitors(instanceName string) []*servicemonitor.Config {

	servicemonitors := make([]*servicemonitor.Config, 0)

	servicemonitors = append(servicemonitors, getSMApplicationController(instanceName))
	servicemonitors = append(servicemonitors, getSMRepoServer(instanceName))
	servicemonitors = append(servicemonitors, getSMServer(instanceName))

	return servicemonitors
}

func getSMServer(instanceName string) *servicemonitor.Config {
	appName := info.GetName()
	monitorlabels := labels.GetMonitorLabels(instanceName, appName)
	ls := labels.GetApplicationLabels(appName)

	// argocd-server
	endpoint := &servicemonitor.ConfigEndpoint{
		Port: "metrics",
		Path: "/metrics",
	}

	ls["app.kubernetes.io/instance"] = "argocd"
	ls["app.kubernetes.io/part-of"] = "argocd"
	ls["app.kubernetes.io/component"] = "server"

	return &servicemonitor.Config{
		Name:                  "argocd-server-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		MonitorMatchingLabels: monitorlabels,
		ServiceMatchingLabels: ls,
		JobName:               "argocd-server",
	}
}

func getSMRepoServer(instanceName string) *servicemonitor.Config {
	appName := info.GetName()
	monitorlabels := labels.GetMonitorLabels(instanceName, appName)
	ls := labels.GetApplicationLabels(appName)

	// argocd-repo-server
	endpoint := &servicemonitor.ConfigEndpoint{
		Port: "metrics",
		Path: "/metrics",
	}

	ls["app.kubernetes.io/instance"] = "argocd"
	ls["app.kubernetes.io/part-of"] = "argocd"
	ls["app.kubernetes.io/component"] = "repo-server"

	return &servicemonitor.Config{
		Name:                  "argocd-repo-server-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		MonitorMatchingLabels: monitorlabels,
		ServiceMatchingLabels: ls,
		JobName:               "argocd-repo-server",
	}

}

func getSMApplicationController(instanceName string) *servicemonitor.Config {
	appName := info.GetName()
	monitorlabels := labels.GetMonitorLabels(instanceName, appName)
	ls := labels.GetApplicationLabels(appName)

	//argocd-application-controller
	endpoint := &servicemonitor.ConfigEndpoint{
		Port: "metrics",
		Path: "/metrics",
	}

	ls["app.kubernetes.io/instance"] = "argocd"
	ls["app.kubernetes.io/part-of"] = "argocd"
	ls["app.kubernetes.io/component"] = "application-controller"

	return &servicemonitor.Config{
		Name:                  "argocd-application-controller-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		MonitorMatchingLabels: monitorlabels,
		ServiceMatchingLabels: ls,
		JobName:               "application-controller",
	}
}
