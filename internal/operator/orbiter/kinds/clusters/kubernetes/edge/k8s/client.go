//go:generate goderive .

package k8s

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/caos/orbiter/internal/helpers"
	"github.com/caos/orbiter/logging"
	"github.com/pkg/errors"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
	mach "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

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
	logger logging.Logger
	set    *kubernetes.Clientset
}

func New(logger logging.Logger, kubeconfig *string) *Client {
	kc := &Client{logger: logger}
	kc.Refresh(kubeconfig)
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
		_, err := resources.Create(rsc)
		return err
	}, func() error {
		_, err := resources.Update(rsc)
		return err
	})
}

func (c *Client) ApplyDeployment(rsc *apps.Deployment) error {
	resources := c.set.AppsV1().Deployments(rsc.GetNamespace())
	return c.apply("deployment", rsc.GetName(), func() error {
		_, err := resources.Create(rsc)
		return err
	}, func() error {
		_, err := resources.Update(rsc)
		return err
	})
}

func (c *Client) ApplySecret(rsc *core.Secret) error {
	resources := c.set.CoreV1().Secrets(rsc.GetNamespace())
	return c.apply("secret", rsc.GetName(), func() error {
		_, err := resources.Create(rsc)
		return err
	}, func() error {
		_, err := resources.Update(rsc)
		return err
	})
}

func (c *Client) ApplyRole(rsc *rbac.Role) error {
	resources := c.set.RbacV1().Roles(rsc.Namespace)
	return c.apply("role", rsc.GetName(), func() error {
		_, err := resources.Create(rsc)
		return err
	}, func() error {
		_, err := resources.Update(rsc)
		return err
	})
}

func (c *Client) ApplyClusterRole(rsc *rbac.ClusterRole) error {
	resources := c.set.RbacV1().ClusterRoles()
	return c.apply("clusterrole", rsc.GetName(), func() error {
		_, err := resources.Create(rsc)
		return err
	}, func() error {
		_, err := resources.Update(rsc)
		return err
	})
}

func (c *Client) ApplyRoleBinding(rsc *rbac.RoleBinding) error {
	resources := c.set.RbacV1().RoleBindings(rsc.Namespace)
	return c.apply("rolebinding", rsc.GetName(), func() error {
		_, err := resources.Create(rsc)
		return err
	}, func() error {
		_, err := resources.Update(rsc)
		return err
	})
}

func (c *Client) ApplyClusterRoleBinding(rsc *rbac.ClusterRoleBinding) error {
	resources := c.set.RbacV1().ClusterRoleBindings()
	return c.apply("clusterrolebinding", rsc.GetName(), func() error {
		_, err := resources.Create(rsc)
		return err
	}, func() error {
		_, err := resources.Update(rsc)
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
	if err == nil || !macherrs.IsNotFound(err) {
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
	return err
}

func (c *Client) GetNode(id string) (node *core.Node, err error) {

	defer func() {
		err = errors.Wrapf(err, "getting node %s failed", id)
	}()

	api, err := c.nodeApi()
	if err != nil {
		return nil, err
	}
	return api.Get(id, mach.GetOptions{})
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

	nodeList, err := api.List(mach.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

func (c *Client) UpdateNode(node *core.Node) (err error) {

	defer func() {
		err = errors.Wrapf(err, "updating node %s failed", node.GetName())
	}()

	api, err := c.nodeApi()
	if err != nil {
		return err
	}
	_, err = api.Update(node)
	return err
}

func (c *Client) cordon(node *core.Node) (err error) {
	defer func() {
		err = errors.Wrapf(err, "cordoning node %s failed", node.GetName())
	}()

	node.Spec.Unschedulable = true
	return c.UpdateNode(node)
}

func (c *Client) Uncordon(node *core.Node) (err error) {
	defer func() {
		err = errors.Wrapf(err, "uncordoning node %s failed", node.GetName())
	}()
	node.Spec.Unschedulable = false
	return c.UpdateNode(node)
}

func (c *Client) Drain(node *core.Node) (err error) {
	defer func() {
		err = errors.Wrapf(err, "draining node %s failed", node.GetName())
	}()

	if err = c.cordon(node); err != nil {
		return err
	}

	return c.evictPods(node)
}

func (c *Client) EnsureDeleted(name string, node NodeWithKubeadm, drain bool) (err error) {

	defer func() {
		err = errors.Wrapf(err, "deleting node %s failed", name)
	}()

	logger := c.logger.WithFields(map[string]interface{}{
		"node": name,
	})
	logger.Debug("Deleting node")

	api, apiErr := c.nodeApi()
	apiErr = errors.Wrap(apiErr, "getting node api failed")

	// Cordon node
	if apiErr == nil {
		nodeStruct, err := api.Get(name, mach.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "getting node %s from kube api failed", name)
		}

		if err = c.Drain(nodeStruct); err != nil {
			return err
		}
	}

	if _, resetErr := node.Execute(nil, nil, "sudo kubeadm reset --force"); resetErr != nil {
		if !strings.Contains(resetErr.Error(), "command not found") {
			return resetErr
		}
	}

	if apiErr != nil {
		return nil
	}
	return api.Delete(name, &mach.DeleteOptions{})
}

func (c *Client) evictPods(node *core.Node) (err error) {

	defer func() {
		err = errors.Wrapf(err, "evicting pods from node %s failed", node.GetName())
	}()

	selector := fmt.Sprintf("spec.nodeName=%s", node.Name)
	podItems, err := c.set.CoreV1().Pods("").List(mach.ListOptions{
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
			logger := c.logger.WithFields(map[string]interface{}{
				"pod":       pod.GetName(),
				"namespace": pod.GetNamespace(),
			})
			logger.Debug("Evicting pod")

			watcher, goErr := c.set.CoreV1().Pods(pod.Namespace).Watch(mach.ListOptions{
				FieldSelector: fmt.Sprintf("metadata.name=%s", pod.Name),
			})
			if goErr != nil {
				synchronizer.Done(nil)
				return
			}
			defer watcher.Stop()

			if goErr := c.set.PolicyV1beta1().Evictions(pod.Namespace).Evict(&policy.Eviction{
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
			logger.Debug("Watching pod")

			timeout := time.After(time.Duration(safeUint64(pod.Spec.TerminationGracePeriodSeconds)) + 30)
			for {
				select {
				case event := <-watcher.ResultChan():
					wPod, ok := event.Object.(*core.Pod)
					if !ok {
						continue
					}
					logger = logger.WithFields(map[string]interface{}{
						"event": event.Type,
					})
					logger.Debug("Pod event happened")
					if event.Type == watch.Deleted {
						logger.WithFields(map[string]interface{}{
							"new_node": wPod.Spec.NodeName,
						}).Debug("Pod evicted")
						synchronizer.Done(nil)
						return
					}
				case <-timeout:
					synchronizer.Done(c.set.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &mach.DeleteOptions{}))
					return
				}
			}
		}(p)
	}
	wg.Wait()

	if synchronizer.IsError() {
		return errors.Wrapf(synchronizer, "concurrently evicting pods from node %s failed", node.Name)
	}
	return nil
}

func safeUint64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func (c *Client) deletePod(pod *core.Pod) error {
	return c.set.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &mach.DeleteOptions{})
}
