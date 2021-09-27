package cmds

import (
	"fmt"

	boomapi "github.com/caos/orbos/v5/internal/operator/boom/api"
	"github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	cmdboom "github.com/caos/orbos/v5/internal/operator/boom/cmd"
	orbnw "github.com/caos/orbos/v5/internal/operator/networking/kinds/orb"
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/git"
	"github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/labels"
)

func deployBoom(monitor mntr.Monitor, gitClient *git.Client, k8sClient kubernetes.ClientInt, binaryVersion string, gitops bool) error {

	if gitops {
		if !gitClient.Exists(git.BoomFile) {
			monitor.Info(fmt.Sprintf("Deployment of BOOM skipped as %s not found in git repo", git.BoomFile))
			return nil
		}
		desiredTree, err := gitClient.ReadTree(git.BoomFile)
		if err != nil {
			return err
		}

		// TODO: Parse toolset in cmdboom.Reconcile function (see deployDatabase, deployNetworking)
		desiredKind, _, apiKind, apiVersion, err := boomapi.ParseToolset(desiredTree)
		if err != nil {
			return err
		}

		if desiredKind != nil &&
			desiredKind.Spec != nil &&
			desiredKind.Spec.Boom != nil &&
			desiredKind.Spec.Boom.Version != "" {
			binaryVersion = desiredKind.Spec.Boom.Version
		}

		if err := cmdboom.Reconcile(
			monitor,
			labels.MustForAPI(labels.MustForOperator("ORBOS", "boom.caos.ch", binaryVersion), apiKind, apiVersion),
			k8sClient,
			desiredKind.Spec.Boom,
			binaryVersion,
			gitops,
		); err != nil {
			return err
		}
	} else {
		boom := &latest.Boom{
			Version:         binaryVersion,
			SelfReconciling: true,
		}

		if err := cmdboom.Reconcile(
			monitor,
			labels.MustForAPI(labels.MustForOperator("ORBOS", "boom.caos.ch", binaryVersion), "BOOM", "v1"),
			k8sClient,
			boom,
			binaryVersion,
			gitops,
		); err != nil {
			return err
		}
	}
	return nil
}

func deployNetworking(monitor mntr.Monitor, gitClient *git.Client, k8sClient kubernetes.ClientInt, version string, gitops bool) error {
	if gitops {
		if !gitClient.Exists(git.NetworkingFile) {
			monitor.Info(fmt.Sprintf("Deployment of networking operator skipped as %s not found in git repo", git.NetworkingFile))
			return nil
		}

		desiredTree, err := gitClient.ReadTree(git.NetworkingFile)
		if err != nil {
			return err
		}
		desired, err := orbnw.ParseDesiredV0(desiredTree)
		if err != nil {
			return err
		}
		spec := desired.Spec

		// at takeoff the artifacts have to be applied
		spec.SelfReconciling = true
		if err := orbnw.Reconcile(
			monitor,
			spec,
			gitops,
		)(k8sClient); err != nil {
			return err
		}
	} else {
		// at takeoff the artifacts have to be applied
		spec := &orbnw.Spec{
			Version:         version,
			SelfReconciling: true,
		}

		if err := orbnw.Reconcile(
			monitor,
			spec,
			gitops,
		)(k8sClient); err != nil {
			return err
		}
	}
	return nil
}
