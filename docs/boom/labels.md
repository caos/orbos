# Labels with BOOM

Different labels are used in the BOOM context to differentiate resources which are managed by BOOM.

## Applications

All resources of the applications deployed by BOOM have several BOOM-specific labels to identify them as managed by BOOM:
- app.kubernetes.io/managed-by: boom.caos.ch
- boom.caos.ch/part-of: boom
- boom.caos.ch/instance: \*instanceName\*
- boom.caos.ch/application: \*application\* 

As example for ambassador:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/instance: ambassador
    app.kubernetes.io/managed-by: boom.caos.ch
    app.kubernetes.io/name: ambassador
    app.kubernetes.io/part-of: ambassador
    boom.caos.ch/application: ambassador
    boom.caos.ch/instance: boom
    boom.caos.ch/part-of: boom
  name: ambassador
  namespace: caos-system
spec:
```

It is important to note that only the resources directly applied from BOOM will get the labels, 
as the resulting pods from the deployments are irrelevant to compare the desired and the current state. 
Because the pods will get created through the replica-sets which are created through Kubernetes from the deployments.

If any other resources have this labels and they are not desired, BOOM will delete the resources without asking,
as this is part of reconciling the toolset.

## ServiceMonitors, PodMonitors and PrometheusRules

The CRD-definitions for Prometheus have their own labels which define if they are included.

Labels that the CRD will be included in the caos-Prometheus-instance:
- boom.caos.ch/prometheus: \*prometheus-instance\*

As example the ServiceMonitor for ambassador:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app.kubernetes.io/managed-by: boom.caos.ch
    boom.caos.ch/application: prometheus
    boom.caos.ch/instance: boom
    boom.caos.ch/part-of: boom
    boom.caos.ch/prometheus: caos
  name: ambassador-servicemonitor
  namespace: caos-system
```

