package servicemonitor

type ConfigEndpoint struct {
	Port              string
	TargetPort        string
	Interval          string
	Scheme            string
	Path              string
	BearerTokenFile   string
	MetricRelabelings []*ConfigRelabeling
	Relabelings       []*ConfigRelabeling
	TLSConfig         *ConfigTLSConfig
	HonorLabels       bool
}

type ConfigTLSConfig struct {
	CaFile     string
	ServerName string
}

type ConfigRelabeling struct {
	Action       string
	Regex        string
	SourceLabels []string
	TargetLabel  string
	Replacement  string
}

type Config struct {
	Name                  string
	Endpoints             []*ConfigEndpoint
	NamespaceSelector     []string
	MonitorMatchingLabels map[string]string
	ServiceMatchingLabels map[string]string
	JobName               string
}

func SpecToValues(config *Config) *Values {

	endpoints := make([]*Endpoint, 0)
	for _, endpoint := range config.Endpoints {
		metricRels := make([]*MetricRelabeling, 0)
		for _, relabel := range endpoint.MetricRelabelings {
			rel := &MetricRelabeling{
				Action:       relabel.Action,
				Regex:        relabel.Regex,
				SourceLabels: relabel.SourceLabels,
				TargetLabel:  relabel.TargetLabel,
				Replacement:  relabel.Replacement,
			}
			metricRels = append(metricRels, rel)
		}
		rels := make([]*Relabeling, 0)
		for _, relabel := range endpoint.Relabelings {
			rel := &Relabeling{
				Action:       relabel.Action,
				Regex:        relabel.Regex,
				SourceLabels: relabel.SourceLabels,
				TargetLabel:  relabel.TargetLabel,
				Replacement:  relabel.Replacement,
			}
			rels = append(rels, rel)
		}
		valueEndpoint := &Endpoint{
			Port:              endpoint.Port,
			TargetPort:        endpoint.TargetPort,
			Interval:          endpoint.Interval,
			Scheme:            endpoint.Scheme,
			Path:              endpoint.Path,
			BearerTokenFile:   endpoint.BearerTokenFile,
			MetricRelabelings: metricRels,
			Relabelings:       rels,
			HonorLabels:       endpoint.HonorLabels,
		}
		if endpoint.TLSConfig != nil {
			t := &TLSConfig{
				CaFile:     endpoint.TLSConfig.CaFile,
				ServerName: endpoint.TLSConfig.ServerName,
			}
			valueEndpoint.TLSConfig = t
		}

		endpoints = append(endpoints, valueEndpoint)
	}

	values := &Values{
		Name:             config.Name,
		AdditionalLabels: config.MonitorMatchingLabels,
		Selector: &Selector{
			MatchLabels: config.ServiceMatchingLabels,
		},
		NamespaceSelector: &NamespaceSelector{
			Any: true,
		},
		JobLabel:  config.JobName,
		Endpoints: endpoints,
	}

	if len(config.NamespaceSelector) != 0 {
		values.NamespaceSelector = &NamespaceSelector{
			Any:        false,
			MatchNames: config.NamespaceSelector,
		}
	}

	return values
}
