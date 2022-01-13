package clientgo

import (
	"context"
	"fmt"
	"strings"

	"github.com/caos/orbos/mntr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

var (
	ignoredResources = []string{
		"componentstatuses",
		"endpoints",
		"bindings",
		"nodes",
		"nodes",
		"replicationcontrollers",
		"podtemplates",
		"limitranges",
		"apiservices",
		"controllerrevisions",
		"leases",
		"backendconfigs",
		"updateinfos",
		"runtimeclasses",
		"storagestates",
		"storageversionmigrations",
		"csidrivers",
		"csinodes",
		"localsubjectaccessreviews",
		"selfsubjectaccessreviews",
		"selfsubjectrulesreviews",
		"subjectaccessreviews",
		"tokenreviews",
		"scalingpolicies",
		"priorityclasses",
	}
	ignoredGroupResources = []string{
		// "metrics.k8s.io/pods",
		"metrics.k8s.io/nodes",
		"networking.gke.io/managedcertificates",
		"networking.gke.io/ingresses",
		"networking.gke.io/networkpolicies",
	}
)

var (
	Limit int64 = 50
)

type ResourceInfo struct {
	Group      string
	Version    string
	Resource   string
	Namespaced bool
}

type Resource struct {
	Group     string
	Version   string
	Resource  string
	Kind      string
	Name      string
	Namespace string
	Labels    map[string]string
}

type ResourceSorter []*Resource

func (a ResourceSorter) Len() int      { return len(a) }
func (a ResourceSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ResourceSorter) Less(i, j int) bool {
	return (a[i].Group < a[j].Group ||
		(a[i].Group == a[j].Group && a[i].Version < a[j].Version) ||
		(a[i].Group == a[j].Group && a[i].Version == a[j].Version && a[i].Resource < a[j].Resource) ||
		(a[i].Group == a[j].Group && a[i].Version == a[j].Version && a[i].Resource == a[j].Resource && a[i].Kind < a[j].Kind) ||
		(a[i].Group == a[j].Group && a[i].Version == a[j].Version && a[i].Resource == a[j].Resource && a[i].Kind == a[j].Kind && a[i].Name < a[j].Name) ||
		(a[i].Group == a[j].Group && a[i].Version == a[j].Version && a[i].Resource == a[j].Resource && a[i].Kind == a[j].Kind && a[i].Name == a[j].Name && a[i].Namespace < a[j].Namespace))
}

func GetResource(monitor mntr.Monitor, group, version, resource, namespace, name string) (*Resource, error) {
	res := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

	conf, err := GetClusterConfig(monitor, "")
	if err != nil {
		return nil, err
	}

	clientset, err := dynamic.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	result, err := clientset.Resource(res).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &Resource{
		Kind:      result.GetKind(),
		Name:      result.GetName(),
		Namespace: result.GetNamespace(),
		Labels:    result.GetLabels(),
	}, nil
}

func DeleteResource(monitor mntr.Monitor, resource *Resource) error {
	res := schema.GroupVersionResource{Group: resource.Group, Version: resource.Version, Resource: resource.Resource}
	conf, err := GetClusterConfig(monitor, "")
	if err != nil {
		return err
	}

	client, err := dynamic.NewForConfig(conf)
	if err != nil {
		return err
	}

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	clientRes := client.Resource(res)
	if resource.Namespace != "" {
		err = clientRes.Namespace(resource.Namespace).Delete(context.Background(), resource.Name, *deleteOptions)
	} else {
		err = clientRes.Delete(context.Background(), resource.Name, *deleteOptions)
	}

	if err != nil {
		return fmt.Errorf("error while deleting %s: %w", resource.Name, err)
	}
	return nil
}

func GetGroupVersionsResources(monitor mntr.Monitor, filtersResources []string) ([]*ResourceInfo, error) {
	listMonitor := monitor.WithFields(map[string]interface{}{
		"action": "groupVersionResources",
	})

	conf, err := GetClusterConfig(monitor, "")
	if err != nil {
		return nil, fmt.Errorf("getting cluster config failed: %w", err)
	}

	client, err := discovery.NewDiscoveryClientForConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("creating discovery client failed: %w", err)
	}

	apiGroups, err := client.ServerGroups()
	if err != nil {
		return nil, fmt.Errorf("getting supported groups and versions failed: %w", err)
	}
	resourceInfoList := make([]*ResourceInfo, 0)
	for _, apiGroup := range apiGroups.Groups {
		version := apiGroup.PreferredVersion
		apiResources, err := client.ServerResourcesForGroupVersion(version.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("getting supported resources failed for %s: %w", version.GroupVersion, err)
		}

		for _, apiResource := range apiResources.APIResources {

			if filtersResources != nil &&
				len(filtersResources) > 0 &&
				containsFilter(filtersResources, apiGroup.Name, version.Version, apiResource.Kind) {
				continue
			}

			resourceInfo := &ResourceInfo{
				Group:      apiGroup.Name,
				Version:    version.Version,
				Resource:   apiResource.Name,
				Namespaced: apiResource.Namespaced,
			}

			groupResource := strings.Join([]string{resourceInfo.Group, resourceInfo.Resource}, "/")
			if !contains(ignoredResources, resourceInfo.Resource) &&
				!contains(ignoredGroupResources, groupResource) {
				resourceInfoList = append(resourceInfoList, resourceInfo)
			}
		}
	}
	listMonitor.WithField("count", len(resourceInfoList)).Debug("Listed groupVersionsResources")
	return resourceInfoList, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ListResources(monitor mntr.Monitor, resourceInfoList []*ResourceInfo, labels map[string]string) ([]*Resource, error) {
	listMonitor := monitor.WithFields(map[string]interface{}{
		"action": "listResources",
	})
	conf, err := GetClusterConfig(monitor, "")
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	labelSelector := ""
	for k, v := range labels {
		if labelSelector == "" {
			labelSelector = strings.Join([]string{k, v}, "=")
		} else {
			keyValue := strings.Join([]string{k, v}, "=")
			labelSelector = strings.Join([]string{labelSelector, keyValue}, ", ")
		}
	}

	listMonitor.WithFields(map[string]interface{}{
		"labelSelector": labelSelector,
	}).Debug(fmt.Sprintf("Used labels"))

	resourceList := make([]*Resource, 0)
	for _, resourceInfo := range resourceInfoList {

		gvr := schema.GroupVersionResource{Group: resourceInfo.Group, Version: resourceInfo.Version, Resource: resourceInfo.Resource}

		listOpt := metav1.ListOptions{
			LabelSelector: labelSelector,
			Limit:         Limit,
		}
		list, err := client.Resource(gvr).List(context.Background(), listOpt)
		if err != nil {
			continue
		}

		resList, err := getResourcesFromList(resourceInfo.Group, resourceInfo.Version, resourceInfo.Resource, list.Items)
		if err != nil {
			return nil, err
		}

		for list.GetContinue() != "" {
			listOpt.Continue = list.GetContinue()
			listInternal, err := client.Resource(gvr).List(context.Background(), listOpt)
			if err != nil {
				continue
			}
			list = listInternal

			resListContinue, err := getResourcesFromList(resourceInfo.Group, resourceInfo.Version, resourceInfo.Resource, list.Items)
			if err != nil {
				return nil, err
			}
			resList = append(resList, resListContinue...)
		}

		listMonitor.WithFields(map[string]interface{}{
			"group":    resourceInfo.Group,
			"version":  resourceInfo.Version,
			"resource": resourceInfo.Resource,
			"count":    len(resList),
		}).Debug("Listed resources")
		resourceList = append(resourceList, resList...)
	}

	listMonitor.WithFields(map[string]interface{}{
		"count": len(resourceList),
	}).Info("All current resources")
	return resourceList, nil
}

func getResourcesFromList(group, version, resource string, list []unstructured.Unstructured) ([]*Resource, error) {

	resourceList := make([]*Resource, 0)
	for _, item := range list {

		name, found, err := unstructured.NestedString(item.Object, "metadata", "name")
		if err != nil || !found {
			return nil, err
		}

		kind, _, err := unstructured.NestedString(item.Object, "kind")
		if err != nil {
			return nil, err
		}

		namespace, _, err := unstructured.NestedString(item.Object, "metadata", "namespace")
		if err != nil {
			return nil, err
		}

		labels, _, err := unstructured.NestedMap(item.Object, "metadata", "labels")
		if err != nil {
			return nil, err
		}

		_, found, err = unstructured.NestedSlice(item.Object, "metadata", "ownerReferences")
		if err != nil {
			return nil, err
		}
		if found == true {
			continue
		}

		labelStrs := make(map[string]string)
		for k, label := range labels {
			labelStrs[k] = label.(string)
		}

		resourceList = append(resourceList, &Resource{
			Group:     group,
			Version:   version,
			Resource:  resource,
			Name:      name,
			Kind:      kind,
			Namespace: namespace,
			Labels:    labelStrs,
		})
	}

	return resourceList, nil
}

func GetFilter(group, version, kind string) string {
	return strings.Join([]string{group, version, kind}, "/")
}

func containsFilter(filters []string, group, version, kind string) bool {
	compFilter := GetFilter(group, version, kind)
	for _, filter := range filters {
		if filter == compFilter {
			return true
		}
	}
	return false
}
