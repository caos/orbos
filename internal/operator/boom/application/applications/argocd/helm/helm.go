package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "argo-cd",
		Version: "1.8.3",
		Index: &chart.Index{
			Name: "argo",
			URL:  "argoproj.github.io/argo-helm",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"argoproj/argocd":             "v1.4.2",
		"quay.io/dexidp/dex":          "v2.14.0",
		"redis":                       "5.0.3",
		"ghcr.io/caos/argocd-secrets": "v1.0.20",
	}
}
