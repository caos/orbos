package metrics

import "github.com/prometheus/client_golang/prometheus"

func SuccessfulWriteCurrentState(url string, path string) {
	labels := prometheus.Labels{
		"url":  url,
		"path": path,
	}
	metrics.gyrCurrentStateWrite.Set(success)

	labels["result"] = "success"
	metrics.currentStateWrite.With(labels).Inc()
}

func FailedWritingCurrentState(url string, path string) {
	labels := prometheus.Labels{
		"url":  url,
		"path": path,
	}
	metrics.gyrCurrentStateWrite.Set(failed)

	labels["result"] = "failure"
	metrics.currentStateWrite.With(labels).Inc()
}

func SuccessfulReadingCurrentState() {
	labels := prometheus.Labels{}
	metrics.gyrCurrentStateRead.Set(success)

	labels["result"] = "success"
	metrics.currentStateRead.With(labels).Inc()
}

func FailedReadingCurrentState() {
	labels := prometheus.Labels{}
	metrics.gyrCurrentStateRead.Set(failed)

	labels["result"] = "failure"
	metrics.currentStateRead.With(labels).Inc()
}
