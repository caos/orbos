package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "argo-cd",
		Version: "2.3.5",
		Index: &chart.Index{
			Name: "argo",
			URL:  "argoproj.github.io/argo-helm",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"argoproj/argocd":    "v1.5.4",
		"quay.io/dexidp/dex": "v2.14.0",
		"redis":              "5.0.3",
		"docker.pkg.github.com/caos/argocd-secrets/argocd": "v1.0.17",
	}
}
