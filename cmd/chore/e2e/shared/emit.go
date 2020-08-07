package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Event struct {
	EventType     string            `json:"event_type,omitempty"`
	ClientPayload map[string]string `json:"client_payload,omitempty"`
	Branch        string            `json:"branch,omitempty"`
}

func Emit(event Event, accessToken, organisation, repository string) error {

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/dispatches", organisation, repository)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github.everest-preview+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("emitting github repository dispatch event at %s failed", url)
	}
	return nil
}
