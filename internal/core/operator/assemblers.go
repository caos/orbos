package operator

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/helpers"
	"github.com/caos/orbiter/logging"
	"gopkg.in/yaml.v2"
)

type Kind struct {
	Kind    string
	Version string
	ID      string
}

type Assembler interface {
	BuildContext() ([]string, func(map[string]interface{}))
	Build(kind map[string]interface{}, nodagentUpdater NodeAgentUpdater, secrets *Secrets, dependantConfig interface{}) (Kind, interface{}, []Assembler, error)
	Ensure(ctx context.Context, secrets *Secrets, ensuredDependencies map[string]interface{}) (interface{}, error)
}

type NodeAgentUpdater func(path []string) *NodeAgentCurrent

type Secrets struct {
	read   func(property string) ([]byte, error)
	write  func(property string, value []byte) error
	delete func(property string) error
}

func (s *Secrets) Read(property string) ([]byte, error) {
	return s.read(property)
}

func (s *Secrets) Write(property string, value []byte) error {
	return s.write(property, value)
}

func (s *Secrets) Delete(property string) error {
	return s.delete(property)
}

type nodeAgentChange struct {
	path []string
	spec *NodeAgentSpec
	curr interface{}
}

type assemblerTree struct {
	path             []string
	node             Assembler
	currentState     interface{}
	kind             Kind
	children         []*assemblerTree
	nodeAgentChanges chan *nodeAgentChange
}

func build(logger logging.Logger, assembler Assembler, desiredSource map[string]interface{}, currentSource map[string]interface{}, secrets *Secrets, dependantConfig interface{}, isRoot bool) (*assemblerTree, error) {
	path, overwrite := assembler.BuildContext()
	assemblerLogger := logger.WithFields(map[string]interface{}{
		"assembler": assembler,
	})
	debugLogger := assemblerLogger.WithFields(map[string]interface{}{
		"path": path,
	})
	var (
		deepDesiredKind map[string]interface{}
		deepCurrentKind map[string]interface{}
		err             error
	)
	if isRoot {
		deepDesiredKind = desiredSource
		deepCurrentKind = currentSource
		if err != nil {
			return nil, err
		}
		path = make([]string, 0)
	} else {
		debugLogger.Debug("Navigating to desiredSource assembler")
		deepDesiredKind, err = drillIn(logger.WithFields(map[string]interface{}{
			"purpose": "build",
			"config":  "spec",
		}), desiredSource, path, false)
		if err != nil {
			return nil, errors.Wrapf(err, "navigating to %s's desired source at path %v failed", assembler, path)
		}
		debugLogger.Debug("Navigating to assembler current state")
		deepCurrentKind, err = drillIn(logger.WithFields(map[string]interface{}{
			"purpose": "build",
			"config":  "spec",
		}), currentSource, path, true)
		if err != nil {
			return nil, errors.Wrapf(err, "navigating to %s's current source at path %v failed", assembler, path)
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
	nodeAgentChanges := make(chan *nodeAgentChange, 1000)
	kind, builtConfig, subassemblers, err := assembler.Build(deepDesiredKind, func(dck map[string]interface{}, naCh chan *nodeAgentChange) func(p []string) *NodeAgentCurrent {
		return func(p []string) *NodeAgentCurrent {
			return newNodeAgentCurrent(assemblerLogger, append([]string{"current"}, p...), dck, naCh)
		}
	}(deepCurrentKind, nodeAgentChanges), secrets, dependantConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "building assembler %s failed", assembler)
	}
	assemblerLogger.Debug("Assembler built")

	tree := &assemblerTree{
		node:             assembler,
		path:             path,
		kind:             kind,
		nodeAgentChanges: nodeAgentChanges,
	}

	tree.children = make([]*assemblerTree, len(subassemblers))
	for idx, subassembler := range subassemblers {
		subTree, err := build(logger, subassembler, deepDesiredKind, deepCurrentKind, secrets, builtConfig, false)
		if err != nil {
			return nil, err
		}
		tree.children[idx] = subTree
	}
	return tree, nil
}

func ensure(ctx context.Context, logger logging.Logger, tree *assemblerTree, secrets *Secrets) (interface{}, error) {
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

func rebuildCurrent(logger logging.Logger, kind map[string]interface{}, tree *assemblerTree) error {
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

		if err := rebuildCurrent(logger, deepKind, subtree); err != nil {
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

	if debugLogger.IsVerbose() {
		debugLogger.Debug("Mapping node agent specs to current state")
		fmt.Println(string(intermediate))
	}

	changesCopy := make(chan *nodeAgentChange, len(tree.nodeAgentChanges))
	close(tree.nodeAgentChanges)
	for newNodeAgent := range tree.nodeAgentChanges {
		changesCopy <- newNodeAgent
		nodeAgent, err := drillIn(logger, currentState, newNodeAgent.path, true)
		if err != nil {
			return errors.Wrapf(err, "navigating to assembler %s's node agent spec at %v in the assemblers current state in order to overwrite it failed", tree.node, newNodeAgent.path)
		}
		nodeAgentCurrentPath := append([]string{"current"}, newNodeAgent.path...)
		nodeAgentCurrent, err := drillIn(logger, deepKind, nodeAgentCurrentPath, true)
		if err != nil {
			return errors.Wrapf(err, "navigating to assembler %s's node agent current at %v in the remote yaml in order to restore it failed", tree.node, nodeAgentCurrentPath)
		}
		nodeAgent["kind"] = "nodeagent.caos.ch/NodeAgent"
		nodeAgent["version"] = "v0"
		nodeAgent["spec"] = newNodeAgent.spec
		nodeAgent["current"] = nodeAgentCurrent["current"]

		if debugLogger.IsVerbose() {
			debugLogger.Debug("Node Agent kind overwritten")
			overwritten, err := yaml.Marshal(deepKind)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(overwritten))
		}
	}

	tree.nodeAgentChanges = changesCopy

	deepKind["kind"] = tree.kind.Kind
	deepKind["version"] = tree.kind.Version
	deepKind["id"] = tree.kind.ID
	deepKind["current"] = currentState

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
