package helpers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/pkg/errors"
)


func Check(url string, status int) (string, error) {
	msg, err := checks.Must(checks.NewHTTPCheck(checks.HTTPCheckConfig{
		CheckName:      "check",
		Timeout:        1 * time.Second,
		URL:            url,
		ExpectedStatus: status,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	})).Execute()
	message := msg.(string)
	return message, errors.Wrap(err, fmt.Sprintf("%s", message))
}
