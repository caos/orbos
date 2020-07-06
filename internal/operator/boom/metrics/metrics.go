package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metricsStruct struct {
	gitClone               *prometheus.CounterVec
	crdFormat              *prometheus.CounterVec
	currentStateWrite      *prometheus.CounterVec
	currentStateRead       *prometheus.CounterVec
	reconcilingBundle      *prometheus.CounterVec
	reconcilingApplication *prometheus.CounterVec
	gyrBoom                prometheus.Gauge
	gyrGit                 prometheus.Gauge
	gyrCurrentStateWrite   prometheus.Gauge
	gyrCurrentStateRead    prometheus.Gauge
	gyrReconciling         prometheus.Gauge
	gyrFormat              prometheus.Gauge
}

var (
	metrics = &metricsStruct{
		gitClone:               newCounterVec("boom_git_clone", "Counter how many times git repositories were cloned", "result", "url"),
		crdFormat:              newCounterVec("boom_crd_format", "Counter how many failures there were with the crd unmarshalling", "result", "url", "reason"),
		currentStateWrite:      newCounterVec("boom_current_state_write", "Counter how many times the current state was written", "result", "url"),
		currentStateRead:       newCounterVec("boom_current_state_read", "Counter how many times the current state was read", "result"),
		reconcilingBundle:      newCounterVec("boom_reconciling_bundle", "Counter how many times the bundle was reconciled", "result", "bundle"),
		reconcilingApplication: newCounterVec("boom_reconciling_application", "Counter how many times a application was reconciled", "result", "application", "templator", "deploy"),
		gyrBoom:                newGauge("boom_gyr", "Status of Boom in GreenYellowRed"),
		gyrGit:                 newGauge("boom_git_gyr", "Status of git connection in GreenYellowRed"),
		gyrCurrentStateWrite:   newGauge("boom_current_state_write_gyr", "Status of current state write in GreenYellowRed"),
		gyrCurrentStateRead:    newGauge("boom_current_state_read_gyr", "Status of current state read in GreenYellowRed"),
		gyrReconciling:         newGauge("boom_reconciling_gyr", "Status of reconciling in GreenYellowRed"),
		gyrFormat:              newGauge("boom_crd_format_gyr", "Status of format unmarshalin in GreenYellowRed"),
	}
	failed  float64 = 0
	success float64 = 1
)

func newGauge(name, help string) prometheus.Gauge {
	gauge := promauto.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})

	err := register(gauge)
	if err != nil {
		return nil
	}

	return gauge
}

func newCounterVec(name string, help string, labels ...string) *prometheus.CounterVec {
	counterVec := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	},
		labels,
	)

	err := register(counterVec)
	if err != nil {
		return nil
	}

	return counterVec
}
func register(collector prometheus.Collector) error {
	err := prometheus.Register(collector)
	_, ok := err.(prometheus.AlreadyRegisteredError)
	if err != nil && !ok {
		return err
	}

	return nil
}
