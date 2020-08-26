package logs

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"strings"

	amlogs "github.com/caos/orbos/internal/operator/boom/application/applications/ambassador/logs"
	aglogs "github.com/caos/orbos/internal/operator/boom/application/applications/argocd/logs"
	glogs "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/logs"
	ksmlogs "github.com/caos/orbos/internal/operator/boom/application/applications/kubestatemetrics/logs"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"
	lologs "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logs"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/info"
	plogs "github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/logs"
	pnelogs "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusnodeexporter/logs"
	pologs "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/logs"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetAllResources(toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec) []interface{} {

	ret := make([]interface{}, 0)
	// output to loki
	if toolsetCRDSpec.LogsPersisting != nil {
		outputNames, outputs := getOutputs(toolsetCRDSpec.LogsPersisting.ClusterOutput)

		// add flows for each application
		flows := getAllFlows(toolsetCRDSpec, outputNames)

		for _, output := range outputs {
			ret = append(ret, output)
		}

		for _, flow := range flows {
			ret = append(ret, flow)
		}
	}

	if len(ret) > 0 {
		//logging resource so that fluentd and fluentbit are deployed
		ret = append(ret, getLogging(toolsetCRDSpec))
	}

	return ret
}

func getLogging(toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec) *logging.Logging {
	conf := &logging.Config{
		Name:             "logging",
		Namespace:        "caos-system",
		ControlNamespace: "caos-system",
	}

	if toolsetCRDSpec.LogCollection.NodeSelector != nil {
		for k, v := range toolsetCRDSpec.LogCollection.NodeSelector {
			conf.NodeSelector[k] = v
		}
	}

	if toolsetCRDSpec.LogCollection.FluentdPVC != nil {
		conf.FluentdPVC = &logging.Storage{
			StorageClassName: toolsetCRDSpec.LogCollection.FluentdPVC.StorageClass,
			Storage:          toolsetCRDSpec.LogCollection.FluentdPVC.Size,
		}
		if toolsetCRDSpec.LogCollection.FluentdPVC.AccessModes != nil {
			conf.FluentdPVC.AccessModes = toolsetCRDSpec.LogCollection.FluentdPVC.AccessModes
		}
	}

	return logging.New(conf)
}

func getAllFlows(toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec, outputNames []string) []*logging.Flow {

	flows := make([]*logging.Flow, 0)
	if toolsetCRDSpec.APIGateway != nil && toolsetCRDSpec.APIGateway.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Ambassador) {
		flows = append(flows, logging.NewFlow(amlogs.GetFlow(outputNames)))
	}

	if toolsetCRDSpec.Monitoring != nil && toolsetCRDSpec.Monitoring.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Grafana) {
		flows = append(flows, logging.NewFlow(glogs.GetFlow(outputNames)))
	}

	if toolsetCRDSpec.MetricCollection != nil && toolsetCRDSpec.MetricCollection.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.PrometheusOperator) {
		flows = append(flows, logging.NewFlow(pologs.GetFlow(outputNames)))
	}

	if toolsetCRDSpec.NodeMetricsExporter != nil && toolsetCRDSpec.NodeMetricsExporter.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.PrometheusNodeExporter) {
		flows = append(flows, logging.NewFlow(pnelogs.GetFlow(outputNames)))
	}

	if toolsetCRDSpec.KubeMetricsExporter != nil && toolsetCRDSpec.KubeMetricsExporter.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.KubeStateMetrics) {
		flows = append(flows, logging.NewFlow(ksmlogs.GetFlow(outputNames)))
	}

	if toolsetCRDSpec.Reconciling != nil && toolsetCRDSpec.Reconciling.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Argocd) {
		flows = append(flows, logging.NewFlow(aglogs.GetFlow(outputNames)))
	}
	if toolsetCRDSpec.LogCollection != nil && toolsetCRDSpec.LogCollection.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.LoggingOperator) {
		flows = append(flows, logging.NewFlow(lologs.GetFlow(outputNames)))
	}

	if toolsetCRDSpec.MetricsPersisting != nil && toolsetCRDSpec.MetricsPersisting.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Prometheus) {
		flows = append(flows, logging.NewFlow(plogs.GetFlow(outputNames)))
	}

	if toolsetCRDSpec.LogsPersisting != nil && toolsetCRDSpec.LogsPersisting.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Loki) {
		flows = append(flows, logging.NewFlow(getLokiFlow(outputNames)))
	}

	return flows
}

func getLokiFlow(outputs []string) *logging.FlowConfig {
	ls := labels.GetApplicationLabels(info.GetName())

	return &logging.FlowConfig{
		Name:         "flow-loki",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}

func getOutputs(clusterOutput bool) ([]string, []*logging.Output) {
	outputURL := strings.Join([]string{"http://", info.GetName().String(), ".", info.GetNamespace(), ":3100"}, "")

	conf := &logging.ConfigOutput{
		Name:      "output-loki",
		Namespace: "caos-system",
		URL:       outputURL,
	}

	outputs := make([]*logging.Output, 0)
	outputs = append(outputs, logging.NewOutput(clusterOutput, conf))
	outputNames := make([]string, 0)
	outputNames = append(outputNames, conf.Name)

	return outputNames, outputs
}
