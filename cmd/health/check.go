package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/pkg/errors"
)

func check(arg string) (string, error) {
	return deriveCompose(
		splitArg,
		parseArg,
		checkParsed)(arg)
}

type splitArgsTuple func() (string, string)

const separator = "@"

func splitArg(arg string) (splitArgsTuple, error) {
	parts := strings.Split(arg, separator)
	if len(parts) < 2 {
		return nil, fmt.Errorf("argument %s should be of form [url]:[expectedstatus], e.g. 200%shttp://httpbin.org/status/200,300", arg, separator)
	}
	return deriveTupleSplitArgs(strings.Join(parts[1:], separator), parts[0]), nil
}

type parsedArgsTuple func() (string, int)

func parseArg(args splitArgsTuple) (parsedArgsTuple, error) {
	url, status := args()
	expectStatus, err := strconv.ParseInt(status, 10, 16)
	if err != nil {
		return nil, err
	}
	return deriveTupleParseArgs(url, int(expectStatus)), nil
}

func checkParsed(args parsedArgsTuple) (string, error) {
	url, status := args()
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
