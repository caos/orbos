package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "argo-cd",
		Version: "2.6.3",
		Index: &chart.Index{
			Name: "argo",
			URL:  "argoproj.github.io/argo-helm",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"argoproj/argocd":    "v1.7.4",
		"quay.io/dexidp/dex": "v2.22.0",
		"redis":              "5.0.8",
		"docker.pkg.github.com/caos/argocd-secrets/argocd": "v1.0.19",
	}
}
