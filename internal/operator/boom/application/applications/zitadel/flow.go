package zitadel

import "github.com/caos/orbos/v5/internal/operator/boom/application/applications/logcollection/logging"

func GetFlows(outputs, clusterOutputs []string) []*logging.FlowConfig {
	return []*logging.FlowConfig{
		getOperatorFlow(outputs, clusterOutputs),
		getZitadelFlow(outputs, clusterOutputs),
	}
}

func getZitadelFlow(outputs, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:           "flow-zitadel",
		Namespace:      "caos-system",
		SelectLabels:   getApplicationServiceLabels(),
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}

func getOperatorFlow(outputs, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:           "flow-zitadel-operator",
		Namespace:      "caos-system",
		SelectLabels:   getOperatorServiceLabels(),
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}
