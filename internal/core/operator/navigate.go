package operator

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/helpers"
	"github.com/caos/orbiter/internal/core/logging"
	"github.com/caos/orbiter/internal/edge/git"
	"gopkg.in/yaml.v2"
)

func toNestedRoot(logger logging.Logger, gitClient *git.Client, path []string, desiredFile string, current map[string]interface{}) (map[string]interface{}, map[string]interface{}, error) {

	workCurrent := current
	var workDesired map[string]interface{}
	if len(path) == 0 {
		var err error
		if workDesired, err = gitClient.Read(desiredFile); err != nil {
			return nil, nil, err
		}
	} else {
		var err error
		if workCurrent, err = drillIn(logger.WithFields(map[string]interface{}{
			"purpose": "navigate to root",
			"config":  "current",
		}), current, path, false); err != nil {
			return nil, nil, err
		}
		workDesired = workCurrent
	}
	immutableDesired := make(map[string]interface{})
	for key, value := range workDesired {
		immutableDesired[key] = value
	}
	workDesired = immutableDesired
	return workDesired, workCurrent, nil
}

func drillIn(logger logging.Logger, cfg map[string]interface{}, path []string, force bool) (map[string]interface{}, error) {
	drillInLogger := logger.WithFields(map[string]interface{}{
		"path":  path,
		"force": force,
	})
	if len(path) == 0 {
		drillInLogger.Debug("Finished drilling in")
		return cfg, nil
	}
	prop := path[0]
	sub, ok := cfg[prop]
	if !ok && !force {
		props := make([]string, 0)
		for p := range cfg {
			props = append(props, p)
		}
		return nil, errors.Errorf("property %s not found in %+v when attempted to drill in", prop, props)
	}
	if force && sub == nil {
		sub = make(map[string]interface{})
	}
	subType, err := helpers.ToStringKeyedMap(sub)
	if err != nil {
		return nil, err
	}
	cfg[prop] = subType

	if logger.IsVerbose() {
		drillInLogger.Debug("Drilled in")
		bytes, err := yaml.Marshal(subType)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bytes))
	}
	return drillIn(logger, subType, path[1:], force)
}
