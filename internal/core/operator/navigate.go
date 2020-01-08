package operator

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/helpers"
	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/logging"
	"gopkg.in/yaml.v2"
)

func toNestedRoot(logger logging.Logger, gitClient *git.Client, path []string, desired map[string]interface{}, current map[string]interface{}) (map[string]interface{}, map[string]interface{}, error) {

	workCurrent := current
	workDesired := desired
	if len(path) > 0 {
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
	if prop == "clusters" || prop == "providers" {
		return drillInSlice(logger, cfg, path, force)
	}

	sub, ok := cfg[prop]
	if !ok && !force {
		return nil, errors.Errorf("drilling in to property %s in structure %+v failed", prop, cfg)
	}

	if !ok && force {
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

func drillInSlice(logger logging.Logger, cfg map[string]interface{}, path []string, force bool) (map[string]interface{}, error) {
	drillInLogger := logger.WithFields(map[string]interface{}{
		"path":  path,
		"force": force,
	})
	if len(path) == 0 || len(path) == 1 {
		drillInLogger.Debug("Finished drilling in")
		return cfg, nil
	}

	prop := path[0]
	id := path[1]
	sliceInterface, ok := cfg[prop]
	if !ok {
		if !force {
			return nil, errors.Errorf("drilling in to property %s in structure %+v failed", prop, cfg)
		}
		newDeps := []map[string]interface{}{{"id": id}}
		cfg[prop] = newDeps
		return drillIn(logger, newDeps[0], path[2:], force)
	}

	slice, ok := sliceInterface.([]interface{})
	if !ok {
		return nil, errors.Errorf("value is not of type slice of interfaces, but of %T", slice)
	}

	for i, itemInterface := range slice {
		item, err := helpers.ToStringKeyedMap(itemInterface)
		if err != nil {
			return nil, err
		}
		slice[i] = item
		cfg[prop] = slice

		itemID, ok := item["id"]
		if !ok {
			return nil, errors.Errorf("dep %+v has no id property", item)
		}

		if itemID == id {
			return drillIn(logger, item, path[2:], force)
		}
	}

	if force {
		newItem := map[string]interface{}{"id": id}
		cfg[prop] = append(slice, newItem)
		return drillIn(logger, newItem, path[2:], force)
	}

	return nil, errors.Errorf("config %+v has no slice with an item of id %s at property %s", cfg, id, prop)
}
