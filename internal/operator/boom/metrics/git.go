package metrics

import "github.com/prometheus/client_golang/prometheus"

func SuccessfulGitClone(url string) {
	labels := prometheus.Labels{
		"url": url,
	}

	metrics.gyrGit.Set(success)

	labels["result"] = "success"
	metrics.gitClone.With(labels).Inc()
}

func FailedGitClone(url string) {
	labels := prometheus.Labels{
		"url": url,
	}

	metrics.gyrGit.Set(failed)

	labels["result"] = "failure"
	metrics.gitClone.With(labels).Inc()
}
