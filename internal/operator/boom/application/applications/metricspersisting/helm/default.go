package helm

import (
	prometheusoperator "github.com/caos/orbos/internal/operator/boom/application/applications/metriccollection/helm"
	"github.com/caos/orbos/pkg/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func DefaultValues(imageTags map[string]string, image string) *Values {
	promValues := &PrometheusValues{
		Enabled: true,
		ServiceAccount: &ServiceAccount{
			Create: true,
		},
		Service: &Service{
			Port:       9090,
			TargetPort: 9090,
			NodePort:   30090,
			Type:       "ClusterIP",
		},
		ServicePerReplica: &ServicePerReplica{
			Enabled:    false,
			Port:       9090,
			TargetPort: 9090,
			NodePort:   30091,
		},
		PodDisruptionBudget: &PodDisruptionBudget{
			Enabled:      false,
			MinAvailable: 1,
		},
		Ingress: &Ingress{
			Enabled: false,
		},
		IngressPerReplica: &IngressPerReplica{
			Enabled: false,
		},
		PodSecurityPolicy: &PodSecurityPolicy{},
		ServiceMonitor: &ServiceMonitor{
			SelfMonitor: false,
		},
		PrometheusSpec: &PrometheusSpec{
			Tolerations:  nil,
			NodeSelector: map[string]string{},
			Image: &Image{
				Repository: image,
				Tag:        imageTags[image],
			},
			RuleSelectorNilUsesHelmValues:           true,
			ServiceMonitorSelectorNilUsesHelmValues: true,
			PodMonitorSelectorNilUsesHelmValues:     true,
			Retention:                               "10d",
			Replicas:                                1,
			LogLevel:                                "info",
			LogFormat:                               "logfmt",
			RoutePrefix:                             "/",
			PodAntiAffinityTopologyKey:              "kubernetes.io/hostname",
			SecurityContext: &SecurityContext{
				RunAsNonRoot: true,
				RunAsUser:    1000,
				FsGroup:      2000,
			},
			RemoteWrite: nil,
			Resources: &k8s.Resources{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("600m"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("300m"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
		},
	}

	return &Values{
		FullnameOverride: "operated",
		DefaultRules: &DefaultRules{
			Create: true,
			Rules: &Rules{
				Alertmanager:                true,
				Etcd:                        true,
				General:                     true,
				K8S:                         true,
				KubeApiserver:               true,
				KubePrometheusNodeAlerting:  true,
				KubePrometheusNodeRecording: true,
				KubernetesAbsent:            true,
				KubernetesApps:              true,
				KubernetesResources:         true,
				KubernetesStorage:           true,
				KubernetesSystem:            true,
				KubeScheduler:               true,
				Network:                     true,
				Node:                        true,
				Prometheus:                  true,
				PrometheusOperator:          true,
				Time:                        true,
			},
		},
		Global: &Global{
			Rbac: &Rbac{
				Create:     true,
				PspEnabled: true,
			},
		},
		Alertmanager: &DisabledToolServicePerReplica{
			Enabled:           false,
			ServicePerReplica: &DisabledTool{Enabled: false},
			IngressPerReplica: &DisabledTool{Enabled: false},
		},
		Grafana: &DisabledTool{
			Enabled: false,
		},
		KubeAPIServer: &DisabledTool{
			Enabled: false,
		},
		Kubelet: &DisabledTool{
			Enabled: false,
		},
		KubeControllerManager: &DisabledTool{
			Enabled: false,
		},
		CoreDNS: &DisabledTool{
			Enabled: false,
		},
		KubeDNS: &DisabledTool{
			Enabled: false,
		},
		KubeEtcd: &DisabledTool{
			Enabled: false,
		},
		KubeScheduler: &DisabledTool{
			Enabled: false,
		},
		KubeProxy: &DisabledTool{
			Enabled: false,
		},
		KubeStateMetricsScrap: &DisabledTool{
			Enabled: false,
		},
		KubeStateMetrics: &DisabledTool{
			Enabled: false,
		},
		NodeExporter: &DisabledTool{
			Enabled: false,
		},
		PrometheusNodeExporter: &DisabledTool{
			Enabled: false,
		},
		PrometheusOperator: &prometheusoperator.PrometheusOperatorValues{
			Enabled: false,
			TLSProxy: &prometheusoperator.TLSProxy{
				Enabled: false,
				Image: &prometheusoperator.Image{
					Repository: "squareup/ghostunnel",
					Tag:        imageTags["squareup/ghostunnel"],
					PullPolicy: "IfNotPresent",
				},
			},
			AdmissionWebhooks: &prometheusoperator.AdmissionWebhooks{
				FailurePolicy: "Fail",
				Enabled:       false,
				Patch: &prometheusoperator.Patch{
					Enabled: false,
					Image: &prometheusoperator.Image{
						Repository: "jettech/kube-webhook-certgen",
						Tag:        imageTags["jettech/kube-webhook-certgen"],
						PullPolicy: "IfNotPresent",
					},
					PriorityClassName: "",
				},
			},
			ServiceAccount: &prometheusoperator.ServiceAccount{
				Create: false,
			},
			ServiceMonitor: &prometheusoperator.ServiceMonitor{
				Interval:    "",
				SelfMonitor: false,
			},
			CreateCustomResource: true,
			KubeletService: &prometheusoperator.KubeletService{
				Enabled: false,
			},
		},
		Prometheus: promValues,
	}
}
