package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

func SuccessfulReconcilingBundle(bundle string) {
	metrics.reconcilingBundle.With(prometheus.Labels{
		"result": "success",
		"bundle": bundle,
	}).Inc()
}

func FailureReconcilingBundle(bundle string) {
	metrics.reconcilingBundle.With(prometheus.Labels{
		"result": "failure",
		"bundle": bundle,
	}).Inc()
}

func SuccessfulReconcilingApplication(app string, templator string, deploy bool) {
	labels := prometheus.Labels{
		"application": app,
		"deploy":      strconv.FormatBool(deploy),
		"templator":   templator,
	}
	metrics.gyrReconciling.Set(success)

	labels["result"] = "success"
	metrics.reconcilingApplication.With(labels).Inc()
}

func FailureReconcilingApplication(app string, templator string, deploy bool) {
	labels := prometheus.Labels{
		"application": app,
		"deploy":      strconv.FormatBool(deploy),
		"templator":   templator,
	}
	metrics.gyrReconciling.Set(failed)

	labels["result"] = "failure"
	metrics.reconcilingApplication.With(labels).Inc()
}
