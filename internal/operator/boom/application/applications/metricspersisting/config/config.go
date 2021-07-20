package config

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/helm"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
)

type Config struct {
	Prefix                  string
	Namespace               string
	MonitorLabels           map[string]string
	RuleLabels              map[string]string
	ServiceMonitors         []*servicemonitor.Config
	ReplicaCount            int
	StorageSpec             *StorageSpec
	AdditionalScrapeConfigs []*helm.AdditionalScrapeConfig
}

type StorageSpec struct {
	StorageClass string
	AccessModes  []string
	Storage      string
}
