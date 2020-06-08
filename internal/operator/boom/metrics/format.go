package metrics

import "github.com/prometheus/client_golang/prometheus"

func WrongCRDFormat(url string, path string) {
	labels := prometheus.Labels{
		"url":    url,
		"path":   path,
		"reason": "structure",
	}
	metrics.gyrFormat.Set(failed)

	labels["reason"] = "structure"
	labels["result"] = "failure"
	metrics.crdFormat.With(labels).Inc()
}

func UnsupportedAPIGroup(url string, path string) {
	labels := prometheus.Labels{
		"url":  url,
		"path": path,
	}
	metrics.gyrFormat.Set(failed)

	labels["result"] = "failure"
	labels["reason"] = "apiGroup"
	metrics.crdFormat.With(labels).Inc()
}

func UnsupportedVersion(url string, path string) {
	labels := prometheus.Labels{
		"url":  url,
		"path": path,
	}
	metrics.gyrFormat.Set(failed)

	labels["result"] = "failure"
	labels["reason"] = "version"
	metrics.crdFormat.With(labels).Inc()
}

func SuccessfulUnmarshalCRD(url string, path string) {
	labels := prometheus.Labels{
		"url":  url,
		"path": path,
	}
	metrics.gyrFormat.Set(success)

	labels["result"] = "success"
	labels["reason"] = "unmarshal"
	metrics.crdFormat.With(labels).Inc()
}
