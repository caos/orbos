package orbiter

/*
import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/helpers"
	"github.com/caos/orbiter/logging"
	"gopkg.in/yaml.v2"
)

type Assembler interface {
	BuildContext() ([]string, func(map[string]interface{}))
	Build(kind map[string]interface{}, nodagentUpdater NodeAgentUpdater, secrets *Secrets, dependantConfig interface{}) (serialized Kind, built interface{}, subassemblers []Assembler, apiVersionCurrent string, err error)
	Ensure(ctx context.Context, secrets *Secrets, ensuredDependencies map[string]interface{}) (interface{}, error)
}

type NodeAgentUpdater func(id string) *NodeAgentCurrent

type nodeAgentChange struct {
	id     string
	mutate func(*NodeAgentSpec)
}

type Desired struct {
	Kind     string
	Version  string
	Spec     interface{} `yaml:",inline"`
	Children []*Desired  `yaml:"deps"`
	ensurer  Assembler   `yaml:"-"`
}

type Secrets struct {
	Node    Node           `yaml:",inline"`
	Secrets yaml.Marshaler `yaml:",inline"`
}

func build(
	logger logging.Logger,
	assembler Assembler,
	desired map[string]interface{},
	secrets map[string]interface{},
	dependantConfig interface{},
	isRoot bool,
	nodeAgentFunc func(id string, changes chan<- *nodeAgentChange) *NodeAgentCurrent) (*assemblerTree, error) {

	path, overwrite := assembler.BuildContext()
	assemblerLogger := logger.WithFields(map[string]interface{}{
		"assembler": assembler,
	})
	debugLogger := assemblerLogger.WithFields(map[string]interface{}{
		"path": path,
	})
	var (
		deepDesiredKind map[string]interface{}
		err             error
	)
	if isRoot {
		deepDesiredKind = desired
		if err != nil {
			return nil, err
		}
		path = make([]string, 0)
	} else {
		debugLogger.Debug("Navigating to desired assembler")
		deepDesiredKind, err = drillIn(logger.WithFields(map[string]interface{}{
			"purpose": "build",
			"config":  "spec",
		}), desired, path, false)
		if err != nil {
			return nil, errors.Wrapf(err, "navigating to %s's desired source at path %v failed", assembler, path)
		}
	}

	if overwrite != nil {
		realDesired, err := helpers.ToStringKeyedMap(deepDesiredKind["spec"])
		if err != nil {
			return nil, errors.Wrapf(err, "converting %s's desired spec %+v to a string keyed map failed", assembler, deepDesiredKind["spec"])
		}
		overwrite(realDesired)
		deepDesiredKind["spec"] = realDesired
	}

	debugLogger.Debug("Building assembler")
	nodeAgentChanges := make(chan *nodeAgentChange)

	kind, builtConfig, subassemblers, builtAPIVersion, err := assembler.Build(deepDesiredKind, func(id string) *NodeAgentCurrent {
		return nodeAgentFunc(id, nodeAgentChanges)
	}, secrets, dependantConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "building assembler %s failed", assembler)
	}
	assemblerLogger.Debug("Assembler built")

	tree := &assemblerTree{
		node:              assembler,
		path:              path,
		kind:              kind,
		currentAPIVersion: builtAPIVersion,
		nodeAgentChanges:  nodeAgentChanges,
	}

	tree.children = make([]*assemblerTree, len(subassemblers))
	for idx, subassembler := range subassemblers {
		subTree, err := build(logger, subassembler, deepDesiredKind, secrets, builtConfig, false, nodeAgentFunc)
		if err != nil {
			return nil, err
		}
		tree.children[idx] = subTree
	}
	return tree, nil
}

func ensure(
	ctx context.Context,
	logger logging.Logger,
	tree *assemblerTree,
	secrets *Secrets) (interface{}, error) {

	assemblerLogger := logger.WithFields(map[string]interface{}{
		"assembler": tree.node,
	})
	debugLogger := assemblerLogger.WithFields(map[string]interface{}{
		"subassemblers": tree.children,
	})
	debugLogger.Debug("Ensuring")
	ensuredChildren := make(map[string]interface{})
	for _, subassembler := range tree.children {
		ensured, err := ensure(ctx, logger, subassembler, secrets)
		if err != nil {
			return nil, err
		}
		ensuredChildren[subassembler.kind.ID] = ensured
	}

	current, err := tree.node.Ensure(ctx, secrets, ensuredChildren)
	if err != nil {
		return current, errors.Wrapf(err, "ensuring assembler %s failed", tree.node)
	}
	tree.currentState = current
	assemblerLogger.Debug("Ensured assemblers desired state")
	return current, nil
}

func desireNodeAgentsFromTree(logger logging.Logger, spec map[string]*NodeAgentSpec, tree *assemblerTree) error {
	debugLogger := logger.WithFields(map[string]interface{}{
		"assembler": tree.node,
		"path":      tree.path,
	})
	debugLogger.Debug("Desiring Node Agents from assembler")

	for change := range tree.nodeAgentChanges {
		nodeagent, ok := spec[change.id]
		if !ok {
			nodeagent = &NodeAgentSpec{}
			spec[change.id] = nodeagent
		}
		change.mutate(nodeagent)
	}

	for _, child := range tree.children {
		if err := desireNodeAgentsFromTree(logger, spec, child); err != nil {
			return err
		}
	}

	return nil
}

func buildDesiredNodeAgents(logger logging.Logger, orbiterCommit string, tree *assemblerTree) ([]byte, error) {

	nodeagents := make(map[string]*NodeAgentSpec)
	if err := desireNodeAgentsFromTree(logger, nodeagents, tree); err != nil {
		return nil, err
	}

	return yaml.Marshal(NodeAgentsDesiredKind{
		NodeAgentsKind: NodeAgentsKind{
			Kind:    "orbiter.caos.ch/NodeAgents",
			Version: "v0",
		},
		Spec: NodeAgentsSpec{
			Commit:     orbiterCommit,
			NodeAgents: nodeagents,
		},
	})
}

func buildCurrent(logger logging.Logger, kind map[string]interface{}, tree *assemblerTree, orbiterCommit string) error {
	debugLogger := logger.WithFields(map[string]interface{}{
		"assembler": tree.node,
		"path":      tree.path,
	})
	debugLogger.Debug("Overwriting current model")

	deepKind, err := drillIn(logger, kind, tree.path, true)
	if err != nil {
		return errors.Wrapf(err, "navigating to assembler %s at %v in order to overwrite its current state failed", tree.node, tree.path)
	}

	//	keepCategories := make([]string, 0)
	for _, subtree := range tree.children {
		//		subtree.currentState

		if err := buildCurrent(logger, deepKind, subtree, orbiterCommit); err != nil {
			return err
		}
	}

	currentState := make(map[string]interface{})
	intermediate, err := yaml.Marshal(tree.currentState)
	if err != nil {
		return errors.Wrapf(err, "marshalling assembler %s's current state %+v in order to overwrite it failed", tree.node, tree.currentState)
	}

	if err := yaml.Unmarshal(intermediate, currentState); err != nil {
		return errors.Wrapf(err, "unmarshalling assembler %s's current state %s in order to overwrite it failed", tree.node, string(intermediate))
	}

	deepKind["kind"] = tree.kind.Kind
	deepKind["currentVersion"] = tree.currentAPIVersion
	if tree.kind.ID != "" {
		deepKind["id"] = tree.kind.ID
	}
	if currentState != nil {
		deepKind["current"] = currentState
	}

	if debugLogger.IsVerbose() {
		debugLogger.Debug("Done overwriting current")
		done, err := yaml.Marshal(deepKind)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(done))
	}

	return nil
}
*/
