package yaml

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"gopkg.in/yaml.v3"
)

func Build(resources *corev1.ResourceRequirements) interface{} {

	ds, err := yaml.Marshal(map[string]interface{}{
		"kind":       "DaemonSet",
		"apiVersion": "apps/v1",
		"metadata": map[string]interface{}{
			"name":      "systemd-exporter",
			"namespace": "caos-system",
			"labels": map[string]string{
				"app": "systemd-exporter",
			},
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"matchLabels": map[string]string{
					"app": "systemd-exporter",
				},
			},
			"updateStrategy": map[string]interface{}{
				"rollingUpdate": map[string]string{
					"maxUnavailable": "100%",
				},
				"type": "RollingUpdate",
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]string{
						"app": "systemd-exporter",
					},
					"annotations": map[string]string{
						"prometheus.io/scrape": "true",
						"prometheus.io/path":   "/metrics",
						"prometheus.io/port":   "9558",
					},
				},
				"spec": map[string]interface{}{
					"tolerations": []map[string]string{{
						"effect":   "NoSchedule",
						"operator": "Exists",
					}},
					"securityContext": map[string]uint8{
						"runAsUser": 0,
					},
					"hostPID": true,
					"containers": []map[string]interface{}{{
						"name":  "systemd-exporter",
						"image": "quay.io/povilasv/systemd_exporter:v0.2.0",
						"securityContext": map[string]bool{
							"privileged": true,
						},
						"args": []string{
							"--log.level=info",
							"--path.procfs=/host/proc",
							"--web.disable-exporter-metrics",
							"--collector.unit-whitelist=kubelet.service|docker.service|node-agentd.service|firewalld.service|keepalived.service|nginx.service|sshd.service",
						},
						"ports": []map[string]interface{}{{
							"name":          "metrics",
							"containerPort": 9558,
							"hostPort":      9558,
						}},
						"volumeMounts": []*volumeMount{{
							Name:      "proc",
							MountPath: "/host/proc",
							ReadOnly:  true,
						}, {
							Name:      "systemd",
							MountPath: "/run/systemd",
							ReadOnly:  true,
						}},
						"resources": resources,
					}},
					"volumes": []*volume{{
						Name: "proc",
						HostPath: hostPath{
							Path: "/proc",
						},
					}, {
						Name: "systemd",
						HostPath: hostPath{
							Path: "/run/systemd",
						},
					}},
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}

	svc, err := yaml.Marshal(map[string]interface{}{
		"kind":       "Service",
		"apiVersion": "v1",
		"metadata": map[string]interface{}{
			"name": "systemd-exporter",
			"labels": map[string]string{
				"app.kubernetes.io/managed-by": "boom.caos.ch",
				"boom.caos.ch/instance":        "boom",
				"boom.caos.ch/part-of":         "boom",
				"boom.caos.ch/prometheus":      "caos",
			},
		},
		"spec": map[string]interface{}{
			"ports": []map[string]interface{}{{
				"name":       "metrics",
				"port":       9558,
				"protocol":   "TCP",
				"targetPort": 9558,
			}},
			"selector": map[string]string{
				"app": "systemd-exporter",
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(`%s
---
%s`, string(ds), string(svc))
}

type volumeMount struct {
	Name      string `yaml:"name,omitempty"`
	MountPath string `yaml:"mountPath,omitempty"`
	ReadOnly  bool   `yaml:"readOnly,omitempty"`
}

type volume struct {
	Name     string   `yaml:"name,omitempty"`
	HostPath hostPath `yaml:"hostPath,omitempty"`
}

type hostPath struct {
	Path string `yaml:"path,omitempty"`
}
