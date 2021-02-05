package k8s

import "encoding/json"

func MarshalYAML(ptr interface{}) (interface{}, error) {
	if ptr == nil {
		return nil, nil
	}

	intermediate, err := json.Marshal(ptr)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	return result, json.Unmarshal(intermediate, &result)
}

func UnmarshalYAML(ptr interface{}, unmarshal func(interface{}) error) error {
	generic := make(map[string]interface{})
	if err := unmarshal(&generic); err != nil {
		return err
	}

	intermediate, err := json.Marshal(generic)
	if err != nil {
		return err
	}

	return json.Unmarshal(intermediate, ptr)
}
