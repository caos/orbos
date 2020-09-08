package logs

import (
	"strings"

	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
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

func GetAllResources(toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) []interface{} {

	// output to loki
	outputNames, clusterOutputNames, outputs := getOutputs(toolsetCRDSpec.Loki.ClusterOutput)

	// add flows for each application
	flows := getAllFlows(toolsetCRDSpec, outputNames, clusterOutputNames)

	ret := make([]interface{}, 0)
	if len(flows) > 0 {

		//logging resource so that fluentd and fluentbit are deployed
		ret = append(ret, getLogging(toolsetCRDSpec))
		for _, output := range outputs {
			ret = append(ret, output)
		}

		for _, flow := range flows {
			ret = append(ret, flow)
		}
	}

	return ret
}

func getLogging(toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) *logging.Logging {

	conf := &logging.Config{
		Name:             "logging",
		Namespace:        "caos-system",
		ControlNamespace: "caos-system",
	}
	if toolsetCRDSpec.LoggingOperator.FluentdPVC != nil {
		conf.FluentdPVC = &logging.Storage{
			StorageClassName: toolsetCRDSpec.LoggingOperator.FluentdPVC.StorageClass,
			Storage:          toolsetCRDSpec.LoggingOperator.FluentdPVC.Size,
		}
		if toolsetCRDSpec.LoggingOperator.FluentdPVC.AccessModes != nil {
			conf.FluentdPVC.AccessModes = toolsetCRDSpec.LoggingOperator.FluentdPVC.AccessModes
		}
	}

	return logging.New(conf)
}

func getAllFlows(toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec, outputNames []string, clusterOutputs []string) []*logging.Flow {

	flows := make([]*logging.Flow, 0)
	if toolsetCRDSpec.Ambassador != nil && toolsetCRDSpec.Ambassador.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.Ambassador) {
		flows = append(flows, logging.NewFlow(amlogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.Grafana != nil && toolsetCRDSpec.Grafana.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.Grafana) {
		flows = append(flows, logging.NewFlow(glogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.PrometheusOperator != nil && toolsetCRDSpec.PrometheusOperator.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.PrometheusOperator) {
		flows = append(flows, logging.NewFlow(pologs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.PrometheusNodeExporter != nil && toolsetCRDSpec.PrometheusNodeExporter.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.PrometheusNodeExporter) {
		flows = append(flows, logging.NewFlow(pnelogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.KubeStateMetrics != nil && toolsetCRDSpec.KubeStateMetrics.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.KubeStateMetrics) {
		flows = append(flows, logging.NewFlow(ksmlogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.Argocd != nil && toolsetCRDSpec.Argocd.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.Argocd) {
		flows = append(flows, logging.NewFlow(aglogs.GetFlow(outputNames, clusterOutputs)))
	}
	if toolsetCRDSpec.LoggingOperator != nil && toolsetCRDSpec.LoggingOperator.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.LoggingOperator) {
		flows = append(flows, logging.NewFlow(lologs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.Prometheus != nil && toolsetCRDSpec.Prometheus.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.Prometheus) {
		flows = append(flows, logging.NewFlow(plogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.Loki != nil && toolsetCRDSpec.Loki.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.Loki) {
		flows = append(flows, logging.NewFlow(getLokiFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.MetricsServer != nil && toolsetCRDSpec.MetricsServer.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.MetricsServer) {
		flows = append(flows, logging.NewFlow(mslogs.GetFlow(outputNames, clusterOutputs)))
	}

	if toolsetCRDSpec.PrometheusSystemdExporter != nil && toolsetCRDSpec.PrometheusSystemdExporter.Deploy &&
		(toolsetCRDSpec.Loki.Logs == nil || toolsetCRDSpec.Loki.Logs.PrometheusSystemdExporter) {
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

func getOutputs(clusterOutput bool) ([]string, []string, []*logging.Output) {
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
