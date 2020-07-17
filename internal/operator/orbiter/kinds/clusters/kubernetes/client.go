//go:generate goderive .

package kubernetes

import (
	"context"
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"strings"
	"sync"
	"time"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"

	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	mach "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clgocore "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type NodeWithKubeadm interface {
	Execute(env map[string]string, stdin io.Reader, cmd string) ([]byte, error)
}

type IDFunc func() string

func (i IDFunc) ID() string {
	return i()
}

type NotAvailableError struct{}

func (n *NotAvailableError) Error() string {
	return "Kubernetes is not available"
}

type Client struct {
	monitor           mntr.Monitor
	set               *kubernetes.Clientset
	dynamic           dynamic.Interface
	apixv1beta1client *apixv1beta1client.ApiextensionsV1beta1Client
	mapper            *restmapper.DeferredDiscoveryRESTMapper
}

func NewK8sClient(monitor mntr.Monitor, kubeconfig *string) *Client {
	kc := &Client{monitor: monitor}
	err := kc.Refresh(kubeconfig)
	if err != nil {
		// do nothing
	}
	return kc
}

func (c *Client) Available() bool {
	return c.set != nil
}

func (c *Client) nodeApi() (clgocore.NodeInterface, error) {
	if c.set == nil {
		return nil, &NotAvailableError{}
	}
	return c.set.CoreV1().Nodes(), nil
}

type File struct {
	Name    string
	Content []byte
}

func (c *Client) ApplyNamespace(rsc *core.Namespace) error {
	resources := c.set.CoreV1().Namespaces()

	return c.apply("namespace", rsc.GetName(), func() error {

		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
		return err
	})
}

func (c *Client) ListNamespaces() (*core.NamespaceList, error) {
	return c.set.CoreV1().Namespaces().List(context.Background(), mach.ListOptions{})
}

func (c *Client) ListSecrets(namespace string, labels map[string]string) (*core.SecretList, error) {
	labelSelector := ""
	for k, v := range labels {
		if labelSelector == "" {
			labelSelector = fmt.Sprintf("%s=%s", k, v)
		} else {
			labelSelector = fmt.Sprintf("%s, %s=%s", labelSelector, k, v)
		}
	}

	return c.set.CoreV1().Secrets(namespace).List(context.Background(), mach.ListOptions{LabelSelector: labelSelector})
}

func (c *Client) ApplyDeployment(rsc *apps.Deployment) error {
	resources := c.set.AppsV1().Deployments(rsc.GetNamespace())
	return c.apply("deployment", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
		return err
	})
}

func (c *Client) WaitUntilDeploymentReady(namespace string, name string) error {
	ctx := context.Background()
	deploy, err := c.set.AppsV1().Deployments(namespace).Get(ctx, name, mach.GetOptions{})
	if err != nil {
		return err
	}

	labelSelector := getLabelSelector(deploy.Spec.Selector.MatchLabels)

	watch, err := c.set.CoreV1().Pods(namespace).Watch(ctx, mach.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return err
	}
	replicas := deploy.Spec.Replicas

	return waitForPodsPhase(watch, int(*replicas), core.PodRunning, true, true)
}

func (c *Client) ApplyService(rsc *core.Service) error {
	resources := c.set.CoreV1().Services(rsc.GetNamespace())
	return c.apply("service", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		svc, err := resources.Get(context.Background(), rsc.Name, mach.GetOptions{})
		if err != nil {
			return err
		}
		rsc.Spec.ClusterIP = svc.Spec.ClusterIP
		rsc.ObjectMeta.ResourceVersion = svc.ObjectMeta.ResourceVersion
		_, err = resources.Update(context.Background(), rsc, mach.UpdateOptions{})
		return err
	})
}

func (c *Client) ApplyJob(rsc *batch.Job) error {
	resources := c.set.BatchV1().Jobs(rsc.Namespace)
	return c.apply("job", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		j, err := resources.Get(context.Background(), rsc.GetName(), mach.GetOptions{})
		if err != nil {
			return err
		}
		if j.GetName() != rsc.GetName() || j.GetNamespace() != rsc.GetNamespace() {
			_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
			return err
		}
		return nil
	})
}

func (c *Client) WaitUntilJobCompleted(namespace string, name string, timeoutSeconds time.Duration) error {
	returnChannel := make(chan error, 1)
	go func() {
		ctx := context.Background()
		job, err := c.set.BatchV1().Jobs(namespace).Get(ctx, name, mach.GetOptions{})
		if err != nil {
			returnChannel <- err
			return
		}

		labelSelector := getLabelSelector(job.Spec.Selector.MatchLabels)

		watch, err := c.set.CoreV1().Pods(namespace).Watch(ctx, mach.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			returnChannel <- err
			return
		}

		returnChannel <- waitForPodsPhase(watch, 1, core.PodSucceeded, false, false)
	}()

	select {
	case res := <-returnChannel:
		return res
	case <-time.After(timeoutSeconds * time.Second):
		return errors.New("timeout while waiting for job to complete")
	}
}

func (c *Client) ApplyPodDisruptionBudget(rsc *policy.PodDisruptionBudget) error {
	resources := c.set.PolicyV1beta1().PodDisruptionBudgets(rsc.Namespace)
	return c.apply("poddisruptionbudget", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		pdb, err := resources.Get(context.Background(), rsc.GetName(), mach.GetOptions{})
		if err != nil {
			return err
		}
		if pdb.GetName() != rsc.GetName() || pdb.GetNamespace() != rsc.GetNamespace() {
			_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
			return err
		}
		return nil
	})
}

func (c *Client) ApplyStatefulSet(rsc *apps.StatefulSet) error {
	resources := c.set.AppsV1().StatefulSets(rsc.Namespace)
	return c.apply("statefulset", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		ss, err := resources.Get(context.Background(), rsc.GetName(), mach.GetOptions{})
		if err != nil {
			return err
		}
		if ss.GetName() != rsc.GetName() || ss.GetNamespace() != rsc.GetNamespace() {
			_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
			return err
		}
		return nil
	})
}

func (c *Client) WaitUntilStatefulsetIsReady(namespace string, name string, containerCheck, readyCheck bool, timeoutSeconds time.Duration) error {
	returnChannel := make(chan error, 1)
	go func() {
		ctx := context.Background()
		sfs, err := c.set.AppsV1().StatefulSets(namespace).Get(ctx, name, mach.GetOptions{})
		if err != nil {
			returnChannel <- err
			return
		}

		labelSelector := getLabelSelector(sfs.Spec.Selector.MatchLabels)

		watch, err := c.set.CoreV1().Pods(namespace).Watch(ctx, mach.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			returnChannel <- err
			return
		}
		replicas := sfs.Spec.Replicas

		returnChannel <- waitForPodsPhase(watch, int(*replicas), core.PodRunning, containerCheck, readyCheck)
	}()

	select {
	case res := <-returnChannel:
		return res
	case <-time.After(timeoutSeconds * time.Second):
		return errors.New("timeout while waiting for job to complete")
	}
}

func (c *Client) DeleteDeployment(namespace, name string) error {
	return c.set.AppsV1().Deployments(namespace).Delete(context.Background(), name, mach.DeleteOptions{})
}

func (c *Client) ApplySecret(rsc *core.Secret) error {
	resources := c.set.CoreV1().Secrets(rsc.GetNamespace())
	return c.apply("secret", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
		return err
	})
}

func (c *Client) WaitForConfigMap(namespace string, name string, timeoutSeconds time.Duration) error {
	returnChannel := make(chan error, 1)
	go func() {
		ctx := context.Background()
		for i := 0; i < int(timeoutSeconds); i++ {
			secret, err := c.set.CoreV1().ConfigMaps(namespace).Get(ctx, name, mach.GetOptions{})
			if err != nil && !macherrs.IsNotFound(err) {
				returnChannel <- err
				return
			} else if secret != nil {
				returnChannel <- nil
				return
			}
			time.Sleep(1)
		}
	}()

	select {
	case res := <-returnChannel:
		return res
	case <-time.After(timeoutSeconds * time.Second):
		return errors.New("timeout while waiting for configmap to be created")
	}
}

func (c *Client) WaitForSecret(namespace string, name string, timeoutSeconds time.Duration) error {
	returnChannel := make(chan error, 1)
	go func() {
		ctx := context.Background()
		for i := 0; i < int(timeoutSeconds); i++ {
			secret, err := c.set.CoreV1().Secrets(namespace).Get(ctx, name, mach.GetOptions{})
			if err != nil && !macherrs.IsNotFound(err) {
				returnChannel <- err
				return
			} else if secret != nil {
				returnChannel <- nil
				return
			}
			time.Sleep(1)
		}
	}()

	select {
	case res := <-returnChannel:
		return res
	case <-time.After(timeoutSeconds * time.Second):
		return errors.New("timeout while waiting for secret to be created")
	}
}

func (c *Client) DeleteSecret(namespace, name string) error {
	return c.set.CoreV1().Secrets(namespace).Delete(context.Background(), name, mach.DeleteOptions{})
}

func (c *Client) ApplyConfigmap(rsc *core.ConfigMap) error {
	resources := c.set.CoreV1().ConfigMaps(rsc.GetNamespace())
	return c.apply("secret", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
		return err
	})
}

func (c *Client) ApplyServiceAccount(rsc *core.ServiceAccount) error {
	resources := c.set.CoreV1().ServiceAccounts(rsc.Namespace)
	return c.apply("serviceaccount", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		sa, err := resources.Get(context.Background(), rsc.GetName(), mach.GetOptions{})
		if err != nil {
			return err
		}
		if sa.GetName() != rsc.GetName() || sa.GetNamespace() != rsc.GetNamespace() {
			_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
			return err
		}
		return nil
	})
}

func (c *Client) ApplyRole(rsc *rbac.Role) error {
	resources := c.set.RbacV1().Roles(rsc.Namespace)
	return c.apply("role", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
		return err
	})
}

func (c *Client) ApplyClusterRole(rsc *rbac.ClusterRole) error {
	resources := c.set.RbacV1().ClusterRoles()
	return c.apply("clusterrole", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
		return err
	})
}

func (c *Client) ApplyRoleBinding(rsc *rbac.RoleBinding) error {
	resources := c.set.RbacV1().RoleBindings(rsc.Namespace)
	return c.apply("rolebinding", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
		return err
	})
}

func (c *Client) ApplyClusterRoleBinding(rsc *rbac.ClusterRoleBinding) error {
	resources := c.set.RbacV1().ClusterRoleBindings()
	return c.apply("clusterrolebinding", rsc.GetName(), func() error {
		_, err := resources.Create(context.Background(), rsc, mach.CreateOptions{})
		return err
	}, func() error {
		_, err := resources.Update(context.Background(), rsc, mach.UpdateOptions{})
		return err
	})
}

func (c *Client) apply(object, name string, create func() error, update func() error) (err error) {
	defer func() {
		err = errors.Wrapf(err, "applying %s %s failed", object, name)
	}()

	if c.set == nil {
		return &NotAvailableError{}
	}

	err = update()
	reason := macherrs.ReasonForError(err)
	if err == nil || (reason != "" && !macherrs.IsNotFound(err)) {
		return err
	}
	return create()
}

func (c *Client) Refresh(kubeconfig *string) (err error) {
	if kubeconfig == nil {
		return
	}

	defer func() {
		err = errors.Wrap(err, "refreshing Kubernetes client failed")
	}()

	clientCfg, err := clientcmd.NewClientConfigFromBytes([]byte(*kubeconfig))
	if err != nil {
		return err
	}

	restCfg, err := clientCfg.ClientConfig()
	if err != nil {
		return err
	}

	c.set, err = kubernetes.NewForConfig(restCfg)
	if err != nil {
		return err
	}

	c.apixv1beta1client, err = apixv1beta1client.NewForConfig(restCfg)

	c.dynamic, err = dynamic.NewForConfig(restCfg)
	if err != nil {
		return err
	}

	dc, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		return err
	}
	c.mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	return err
}

func (c *Client) RefreshConfig(config *rest.Config) (err error) {
	c.set, err = kubernetes.NewForConfig(config)
	return err
}

func (c *Client) GetNode(id string) (node *core.Node, err error) {

	api, err := c.nodeApi()
	if err != nil {
		return nil, err
	}
	return api.Get(context.Background(), id, mach.GetOptions{})
}

func (c *Client) ListNodes(filterID ...string) (nodes []core.Node, err error) {
	defer func() {
		err = errors.Wrapf(err, "listing nodes %s failed", strings.Join(filterID, ", "))
	}()

	api, err := c.nodeApi()
	if err != nil {
		return nil, err
	}

	labelSelector := ""
	for _, id := range filterID {
		labelSelector = fmt.Sprintf("%s,name=%s", labelSelector, id)
	}
	if len(labelSelector) > 0 {
		labelSelector = labelSelector[1:]
	}

	nodeList, err := api.List(context.Background(), mach.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

func (c *Client) updateNode(node *core.Node) (err error) {

	defer func() {
		err = errors.Wrapf(err, "updating node %s failed", node.GetName())
	}()

	api, err := c.nodeApi()
	if err != nil {
		return err
	}
	_, err = api.Update(context.Background(), node, mach.UpdateOptions{})
	return err
}

func (c *Client) cordon(node *core.Node) (err error) {
	defer func() {
		err = errors.Wrapf(err, "cordoning node %s failed", node.GetName())
	}()

	monitor := c.monitor.WithFields(map[string]interface{}{
		"machine": node.GetName(),
	})
	monitor.Info("Cordoning node")

	node.Spec.Unschedulable = true
	if err := c.updateNode(node); err != nil {
		return err
	}
	monitor.Info("Node cordoned")
	return nil
}

func (c *Client) Uncordon(machine *Machine, node *core.Node) (err error) {
	defer func() {
		err = errors.Wrapf(err, "uncordoning node %s failed", node.GetName())
	}()
	monitor := c.monitor.WithFields(map[string]interface{}{
		"machine": node.GetName(),
	})
	monitor.Info("Uncordoning node")

	node.Spec.Unschedulable = false
	if err := c.updateNode(node); err != nil {
		return err
	}
	if !machine.Online {
		machine.Online = true
		monitor.Changed("Node uncordoned")
	}
	return nil
}

func (c *Client) Drain(machine *Machine, node *core.Node) (err error) {
	defer func() {
		err = errors.Wrapf(err, "draining node %s failed", node.GetName())
	}()

	monitor := c.monitor.WithFields(map[string]interface{}{
		"machine": node.GetName(),
	})
	monitor.Info("Draining node")

	if err = c.cordon(node); err != nil {
		return err
	}

	if err := c.evictPods(node); err != nil {
		return err
	}
	if machine.Online {
		machine.Online = false
		monitor.Changed("Node drained")
	}
	return nil
}

func (c *Client) EnsureDeleted(name string, machine *Machine, node NodeWithKubeadm, drain bool) (err error) {

	defer func() {
		err = errors.Wrapf(err, "deleting node %s failed", name)
	}()

	monitor := c.monitor.WithFields(map[string]interface{}{
		"machine": name,
	})
	monitor.Info("Ensuring node is deleted")

	api, apiErr := c.nodeApi()
	apiErr = errors.Wrap(apiErr, "getting node api failed")

	// Cordon node
	if apiErr == nil {
		nodeStruct, err := api.Get(context.Background(), name, mach.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "getting node %s from kube api failed", name)
		}

		if err = c.Drain(machine, nodeStruct); err != nil {
			return err
		}
	}

	monitor.Info("Resetting kubeadm")
	if _, resetErr := node.Execute(nil, nil, "sudo kubeadm reset --force"); resetErr != nil {
		if !strings.Contains(resetErr.Error(), "command not found") {
			return resetErr
		}
	}

	if apiErr != nil {
		return nil
	}
	monitor.Info("Deleting node")
	if err := api.Delete(context.Background(), name, mach.DeleteOptions{}); err != nil {
		return err
	}
	if machine.Online || machine.Joined {
		machine.Online = false
		machine.Joined = false
		monitor.Changed("Node deleted")
	}
	return nil
}

func (c *Client) evictPods(node *core.Node) (err error) {

	defer func() {
		err = errors.Wrapf(err, "evicting pods from node %s failed", node.GetName())
	}()

	monitor := c.monitor.WithFields(map[string]interface{}{
		"machine": node.GetName(),
	})

	monitor.Info("Evicting pods")

	selector := fmt.Sprintf("spec.nodeName=%s,status.phase=Running", node.Name)
	podItems, err := c.set.CoreV1().Pods("").List(context.Background(), mach.ListOptions{
		FieldSelector: selector,
	})
	if err != nil {
		return errors.Wrapf(err, "listing pods with selector %s failed", selector)
	}

	// --ignore-daemonsets
	pods := deriveFilter(func(pod core.Pod) bool {
		controllerRef := mach.GetControllerOf(&pod)
		return controllerRef == nil || controllerRef.Kind != apps.SchemeGroupVersion.WithKind("DaemonSet").Kind
	}, append([]core.Pod{}, podItems.Items...))

	var wg sync.WaitGroup
	synchronizer := helpers.NewSynchronizer(&wg)

	for _, p := range pods {
		wg.Add(1)
		go func(pod core.Pod) {

			var gracePeriodSeconds int64 = 60
			monitor := c.monitor.WithFields(map[string]interface{}{
				"pod":       pod.GetName(),
				"namespace": pod.GetNamespace(),
			})
			monitor.Debug("Evicting pod")

			watcher, goErr := c.set.CoreV1().Pods(pod.Namespace).Watch(context.Background(), mach.ListOptions{
				FieldSelector: fmt.Sprintf("metadata.name=%s", pod.Name),
			})
			if goErr != nil {
				synchronizer.Done(nil)
				return
			}
			defer watcher.Stop()

			if goErr := c.set.PolicyV1beta1().Evictions(pod.Namespace).Evict(context.Background(), &policy.Eviction{
				TypeMeta: mach.TypeMeta{
					Kind:       "EvictionKind",
					APIVersion: c.set.PolicyV1beta1().RESTClient().APIVersion().String(),
				},
				ObjectMeta: mach.ObjectMeta{
					Name:      pod.Name,
					Namespace: pod.Namespace,
				},
				DeleteOptions: &mach.DeleteOptions{
					GracePeriodSeconds: &gracePeriodSeconds,
				},
			}); goErr != nil {
				synchronizer.Done(errors.Wrapf(goErr, "evicting pod %s failed", pod.Name))
				return
			}
			monitor.Debug("Watching pod")

			timeout := time.After(time.Duration(safeUint64(pod.Spec.TerminationGracePeriodSeconds)) + 30)
			for {
				select {
				case event := <-watcher.ResultChan():
					wPod, ok := event.Object.(*core.Pod)
					if !ok {
						continue
					}
					monitor = monitor.WithFields(map[string]interface{}{
						"event": event.Type,
					})
					monitor.Debug("Pod event happened")
					if event.Type == watch.Deleted {
						monitor.WithFields(map[string]interface{}{
							"new_node": wPod.Spec.NodeName,
						}).Debug("Pod evicted")
						synchronizer.Done(nil)
						return
					}
				case <-timeout:
					synchronizer.Done(errors.Wrapf(c.set.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, mach.DeleteOptions{}), "Deleting pod %s after timout exceeded failed", pod.Name))
					return
				}
			}
		}(p)
	}
	wg.Wait()

	if synchronizer.IsError() {
		return errors.Wrapf(synchronizer, "concurrently evicting pods from node %s failed", node.Name)
	}

	monitor.Info("Pods evicted")
	return nil
}

func safeUint64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func (c *Client) deletePod(pod *core.Pod) error {
	return c.set.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, mach.DeleteOptions{})
}

func getLabelSelector(labels map[string]string) string {
	labelSelector := ""
	for k, v := range labels {
		if labelSelector == "" {
			labelSelector = k + "=" + v
		} else {
			labelSelector = labelSelector + ", " + k + "=" + v

		}
	}
	return labelSelector
}

type podStatus struct {
	name              string
	phase             core.PodPhase
	containerStatuses []containerStatus
}
type containerStatus struct {
	containerState core.ContainerStatus
}

func waitForPodsPhase(watch watch.Interface, replicaCount int, phase core.PodPhase, containerCheck bool, readyCheck bool) error {
	statusChan := make(chan podStatus)
	pods := make(map[string]core.PodPhase, 0)

	if watch == nil {
		return errors.New("no pods watched")
	}

	go func() {
		for event := range watch.ResultChan() {
			p, ok := event.Object.(*core.Pod)
			if !ok {
				continue
			}
			containerStatuses := make([]containerStatus, 0)
			for _, podContainerStatus := range p.Status.ContainerStatuses {
				containerStatuses = append(containerStatuses, containerStatus{
					containerState: podContainerStatus,
				})
			}

			statusChan <- podStatus{p.Name, p.Status.Phase, containerStatuses}

		}
	}()

	runningPods := 0
	for runningPods < replicaCount {
		podStat := <-statusChan
		if podStat.phase == phase {
			running := true
			ready := true
			if containerCheck {
				running = checkContainer(podStat.containerStatuses)
			}
			if readyCheck {
				ready = checkReady(podStat.containerStatuses)
			}

			if running && ready {
				pods[podStat.name] = podStat.phase
			} else {
				delete(pods, podStat.name)
			}
		} else {
			delete(pods, podStat.name)
		}
		runningPods = len(pods)
	}
	return nil
}

func checkContainer(containers []containerStatus) bool {
	running := true
	for _, stat := range containers {
		if stat.containerState.State.Running == nil {
			running = false
		}
	}
	return running
}

func checkReady(containers []containerStatus) bool {
	ready := true
	for _, stat := range containers {
		if !stat.containerState.Ready {
			ready = false
		}
	}
	return ready
}

func (c *Client) CheckCRD(name string) (*apixv1beta1.CustomResourceDefinition, error) {
	crds := c.apixv1beta1client.CustomResourceDefinitions()

	return crds.Get(context.Background(), name, mach.GetOptions{})
}

func (c *Client) GetNamespacedCRDResource(group, version, kind, namespace, name string) (*unstructured.Unstructured, error) {
	mapping, err := c.mapper.RESTMapping(schema.GroupKind{
		Group: group,
		Kind:  kind,
	}, version)
	if err != nil {
		return nil, err
	}
	resource := c.dynamic.Resource(mapping.Resource).Namespace(namespace)

	return resource.Get(context.Background(), name, mach.GetOptions{})
}

func (c *Client) ApplyNamespacedCRDResource(group, version, kind, namespace, name string, crd *unstructured.Unstructured) error {
	mapping, err := c.mapper.RESTMapping(schema.GroupKind{
		Group: group,
		Kind:  kind,
	}, version)
	if err != nil {
		return err
	}

	resources := c.dynamic.Resource(mapping.Resource).Namespace(namespace)

	return c.apply("crd", name, func() error {
		_, err := resources.Create(context.Background(), crd, mach.CreateOptions{})
		return err
	}, func() error {
		_, err := resources.Update(context.Background(), crd, mach.UpdateOptions{})
		return err
	})
}
