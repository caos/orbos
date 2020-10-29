package logs

import (
	"strings"

	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	amlogs "github.com/caos/orbos/internal/operator/boom/application/applications/ambassador/logs"

	aglogs "github.com/caos/orbos/internal/operator/boom/application/applications/argocd/logs"
	glogs "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/logs"
	ksmlogs "github.com/caos/orbos/internal/operator/boom/application/applications/kubestatemetrics/logs"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"
	lologs "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logs"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/info"
	mslogs "github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver/logs"
	plogs "github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/logs"
	pnelogs "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusnodeexporter/logs"
	pologs "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/logs"
	pselogs "github.com/caos/orbos/internal/operator/boom/application/applications/prometheussystemdexporter/logs"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetAllResources(toolsetCRDSpec *toolsetslatest.ToolsetSpec) []interface{} {

	if toolsetCRDSpec.LogCollection == nil || !toolsetCRDSpec.LogCollection.Deploy {
		return nil
	}

	ret := []interface{}{logging.New(toolsetCRDSpec.LogCollection)}
	// output to loki
	if toolsetCRDSpec.LogsPersisting != nil && toolsetCRDSpec.LogsPersisting.Deploy {
		outputNames, clusterOutputNames, outputs := getLokiOutput(toolsetCRDSpec.LogsPersisting.ClusterOutput)

		// add flows for each application
		flows := getAllFlows(toolsetCRDSpec, outputNames, clusterOutputNames)

		for _, output := range outputs {
			ret = append(ret, output)
		}

		for _, flow := range flows {
			ret = append(ret, flow)
		}
	}

	return ret
}

func getAllFlows(toolsetCRDSpec *toolsetslatest.ToolsetSpec, outputNames []string, clusterOutputs []string) []*logging.Flow {

	flows := make([]*logging.Flow, 0)
	if toolsetCRDSpec.APIGateway != nil && toolsetCRDSpec.APIGateway.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Ambassador) {
		flows = append(flows, logging.NewFlow(amlogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.Monitoring != nil && toolsetCRDSpec.Monitoring.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Grafana) {
		flows = append(flows, logging.NewFlow(glogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.MetricCollection != nil && toolsetCRDSpec.MetricCollection.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.PrometheusOperator) {
		flows = append(flows, logging.NewFlow(pologs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.NodeMetricsExporter != nil && toolsetCRDSpec.NodeMetricsExporter.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.PrometheusNodeExporter) {
		flows = append(flows, logging.NewFlow(pnelogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.KubeMetricsExporter != nil && toolsetCRDSpec.KubeMetricsExporter.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.KubeStateMetrics) {
		flows = append(flows, logging.NewFlow(ksmlogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.Reconciling != nil && toolsetCRDSpec.Reconciling.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Argocd) {
		flows = append(flows, logging.NewFlow(aglogs.GetFlow(outputNames, clusterOutputs)))
	}
	if toolsetCRDSpec.LogCollection != nil && toolsetCRDSpec.LogCollection.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.LoggingOperator) {
		flows = append(flows, logging.NewFlow(lologs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.MetricsPersisting != nil && toolsetCRDSpec.MetricsPersisting.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Prometheus) {
		flows = append(flows, logging.NewFlow(plogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.LogsPersisting != nil && toolsetCRDSpec.LogsPersisting.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.Loki) {
		flows = append(flows, logging.NewFlow(getLokiFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.MetricsServer != nil && toolsetCRDSpec.MetricsServer.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.MetricsServer) {
		flows = append(flows, logging.NewFlow(mslogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.SystemdMetricsExporter != nil && toolsetCRDSpec.SystemdMetricsExporter.Deploy &&
		(toolsetCRDSpec.LogsPersisting.Logs == nil || toolsetCRDSpec.LogsPersisting.Logs.PrometheusSystemdExporter) {
		flows = append(flows, logging.NewFlow(pselogs.GetFlow(outputNames, clusterOutputs)))
	}

	return flows
}

func getLokiFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	ls := labels.GetApplicationLabels(info.GetName())

	return &logging.FlowConfig{
		Name:           "flow-loki",
		Namespace:      "caos-system",
		SelectLabels:   ls,
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}

func getLokiOutput(clusterOutput bool) ([]string, []string, []*logging.Output) {
	outputURL := strings.Join([]string{"http://", info.GetName().String(), ".", info.GetNamespace(), ":3100"}, "")

	conf := &logging.ConfigOutput{
		Name:      "output-loki",
		Namespace: "caos-system",
		URL:       outputURL,
	}

	outputs := make([]*logging.Output, 0)
	outputs = append(outputs, logging.NewOutput(clusterOutput, conf))

	outputNames := make([]string, 0)
	clusterOutputs := make([]string, 0)
	if clusterOutput {
		clusterOutputs = append(clusterOutputs, conf.Name)
	} else {
		outputNames = append(outputNames, conf.Name)
	}

	return outputNames, clusterOutputs, outputs
}
