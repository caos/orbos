package metrics

import "github.com/prometheus/client_golang/prometheus"

func WrongCRDFormat() {
	labels := prometheus.Labels{
		"reason": "structure",
		"result": "failure",
	}
	metrics.gyrFormat.Set(failed)
	metrics.crdFormat.With(labels).Inc()
}

func UnsupportedAPIGroup() {
	labels := prometheus.Labels{
		"reason": "apiGroup",
		"result": "failure",
	}
	metrics.gyrFormat.Set(failed)
	metrics.crdFormat.With(labels).Inc()
}

func UnsupportedVersion() {
	labels := prometheus.Labels{
		"reason": "version",
		"result": "failure",
	}
	metrics.gyrFormat.Set(failed)
	metrics.crdFormat.With(labels).Inc()
}

func SuccessfulUnmarshalCRD() {
	labels := prometheus.Labels{
		"reason": "unmarshal",
		"result": "success",
	}
	metrics.gyrFormat.Set(success)
	metrics.crdFormat.With(labels).Inc()
}
