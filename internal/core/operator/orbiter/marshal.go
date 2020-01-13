package orbiter

/*
import (
	"sort"

	"github.com/caos/orbiter/internal/core/helpers"
	"gopkg.in/yaml.v2"
)

func Marshal(content map[string]interface{}) ([]byte, error) {
	unorderedSerialized, err := yaml.Marshal(content)
	if err != nil {
		return nil, err
	}

	interfacesMap := make(map[string]interface{})
	if err = yaml.Unmarshal(unorderedSerialized, interfacesMap); err != nil {
		return nil, err
	}

	return yaml.Marshal(deeplyConvert(interfacesMap))
}

type node map[string]interface{}

func (n *node) MarshalYAML() ([]byte, error) {

	mapItems := make([]yaml.MapItem, 0)
	for key, value := range *n {
		mapItems = append(mapItems, yaml.MapItem{
			Key:   key,
			Value: value,
		})
	}

	mapSlice := yaml.MapSlice(mapItems)
	sort.Sort(byKey(mapSlice))

	return yaml.Marshal(mapSlice)
}

type byKey yaml.MapSlice

func (a byKey) Len() int           { return len(a) }
func (a byKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byKey) Less(i, j int) bool { return a[i].Key.(string) < a[j].Key.(string) }

func deeplyConvert(tree map[string]interface{}) node {

	for key, value := range tree {
		subTree, err := helpers.ToStringKeyedMap(value)
		if err != nil {
			continue
		}
		tree[key] = deeplyConvert(subTree)
	}

	return node(tree)
}
*/
