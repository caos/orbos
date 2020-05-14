package helpers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/pkg/errors"
)

func Check(url string, status int) (string, error) {

	urlParts := strings.Split(url, "://")
	if urlParts[0] == "tcp" {
		return check(checks.NewPingCheck("tcp", checks.NewDialPinger("tcp", urlParts[1]), 2*time.Second))
	}

	return check(checks.NewHTTPCheck(checks.HTTPCheckConfig{
		CheckName:      "http",
		Timeout:        1 * time.Second,
		URL:            url,
		ExpectedStatus: status,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}))
}

func check(check checks.Check, err error) (string, error) {
	msg, err := checks.Must(check, err).Execute()
	message, ok := msg.(string)
	if !ok {
		return "", err
	}
	return message, errors.Wrap(err, fmt.Sprintf("%s", message))
}
