package crd

import (
	"errors"
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/bundle"
	bundleconfig "github.com/caos/orbos/internal/operator/boom/bundle/config"
	"github.com/caos/orbos/internal/operator/boom/crd/config"
	"github.com/caos/orbos/internal/operator/boom/metrics"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/mntr"
)

const (
	version name.Version = "v1beta2"
)

type Crd struct {
	bundle  *bundle.Bundle
	monitor mntr.Monitor
	status  error
}

func (c *Crd) GetStatus() error {
	return c.status
}

func (c *Crd) SetBackStatus() {
	c.status = nil
}

func (c *Crd) CleanUp() {
	if c.GetStatus() != nil {
		return
	}

	c.status = c.bundle.CleanUp()
}

func GetVersion() name.Version {
	return version
}

func New(conf *config.Config) *Crd {
	crdMonitor := conf.Monitor.WithFields(map[string]interface{}{
		"version": GetVersion(),
	})

	return &Crd{
		monitor: crdMonitor,
		status:  nil,
	}
}

func (c *Crd) SetBundle(conf *bundleconfig.Config) {
	if c.GetStatus() != nil {
		return
	}
	bundle := bundle.New(conf)

	c.status = bundle.AddApplicationsByBundleName(conf.BundleName)
	if c.status != nil {
		return
	}

	c.bundle = bundle
}

func (c *Crd) GetBundle() *bundle.Bundle {
	return c.bundle
}

//
//func (c *Crd) ReconcileWithFunc(currentResourceList []*clientgo.Resource, getToolsetCRD func(instance runtime.Object) error) {
//	if c.GetStatus() != nil {
//		return
//	}
//
//	if getToolsetCRD == nil {
//		c.status = errors.New("ToolsetCRDFunc is nil")
//		c.monitor.Error(c.status)
//		return
//	}
//
//	var toolsetCRD *toolsetsv1beta1.Toolset
//	if err := getToolsetCRD(toolsetCRD); err != nil {
//		c.status = err
//		return
//	}
//
//	c.Reconcile(currentResourceList, toolsetCRD)
//}

func (c *Crd) Reconcile(currentResourceList []*clientgo.Resource, toolsetCRD *toolsetsv1beta2.Toolset) {
	if c.GetStatus() != nil {
		return
	}

	logFields := map[string]interface{}{
		"CRD":    toolsetCRD.Metadata.Name,
		"action": "reconciling",
	}
	monitor := c.monitor.WithFields(logFields)

	if toolsetCRD == nil {
		c.status = errors.New("ToolsetCRD is nil")
		monitor.Error(c.status)
		return
	}

	if c.bundle == nil {
		c.status = errors.New("No bundle for crd")
		monitor.Error(c.status)
		return
	}

	c.status = c.bundle.Reconcile(currentResourceList, toolsetCRD.Spec)
	if c.status != nil {
		metrics.FailureReconcilingBundle(c.bundle.GetPredefinedBundle())
		return
	}
	metrics.SuccessfulReconcilingBundle(c.bundle.GetPredefinedBundle())
}
