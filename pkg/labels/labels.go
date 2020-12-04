package labels

import (
	"gopkg.in/yaml.v3"
)

type Labels interface {
	comparable
	yaml.Marshaler
	Major() int8
	//	internalModel() internalModel
}

/*
type internalModel interface {
	major() int8
}
*/
type comparable interface {
	Equal(comparable) bool
}

func K8sMap(l Labels) (map[string]string, error) {
	return toMapOfStrings(l)
}

func MustK8sMap(l Labels) map[string]string {
	m, err := K8sMap(l)
	if err != nil {
		panic(err)
	}
	return m
}

/*
func MustBreak(old Labels, new Labels) {

	newModel := new.internalModel()
	oldModel := old.internalModel()

	if newModel.major() <= oldModel.major() {
		fmt.Printf("old labels: %s\n", mustToString(oldModel))
		fmt.Printf("new labels: %s\n", mustToString(newModel))
		panic("labels are not breaking")
	}
}
*/
func toMapOfStrings(sth interface{}) (map[string]string, error) {
	someBytes, err := yaml.Marshal(sth)
	if err != nil {
		return nil, err
	}
	mapOfStrings := make(map[string]string)
	return mapOfStrings, yaml.Unmarshal(someBytes, mapOfStrings)
}

/*
func mustToString(model internalModel) string {

	m, err := toMapOfStrings(model)
	if err != nil {
		panic(err)
	}

	var out string
	for key, value := range m {
		out = fmt.Sprintf("%s %s=%s", out, key, value)
	}
	return strings.TrimSpace(out)
}
*/
