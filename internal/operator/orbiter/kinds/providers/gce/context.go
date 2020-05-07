package gce

import (
	"encoding/json"

	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

func clientContext(jsonKey []byte) (string, option.ClientOption, error) {
	key := struct {
		ProjectID string `json:"project_id"`
	}{}
	return key.ProjectID, option.WithCredentialsJSON(jsonKey), errors.Wrap(json.Unmarshal(jsonKey, &key), "extracting project id from jsonkey failed")
}
