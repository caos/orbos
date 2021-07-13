package mntr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

// need to be global variables so that clients can be removed too
var (
	sentryClient        *sentry.Client
	env, dsn, comp, rel string
	doIngest            bool
)

func Ingest(monitor Monitor, version, commit, component, environment string) error {
	if rel != "" || dsn != "" {
		panic("Ingest was already called")
	}
	if version == "" || commit == "" {
		panic("version, commit and dsn must not be empty")
	}
	rel = fmt.Sprintf("%s-%s", version, commit)
	comp = component
	env = environment
	doIngest = true

	go func() {
		for range time.NewTicker(15 * time.Minute).C {
			monitor.Error(fetchDSN())
		}
	}()

	return fetchDSN()
}

func fetchDSN() error {

	resp, err := http.Get("https://raw.githubusercontent.com/caos/sentry-dsns/main/" + comp)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dsn = string(body)
	configure()
	return nil
}

func SwitchEnvironment(environment string) {
	if environment == "" {
		panic("environment must not be empty")
	}
	env = environment

	configure()
}

func Environment() (string, bool) {
	return env, doIngest
}

func configure() {

	if sentryClient != nil {
		sentryClient.Flush(time.Second * 2)
	}

	if env == "" || rel == "" {
		panic(errors.New("call mntr.Ingest first"))
	}

	var err error
	sentryClient, err = sentry.NewClient(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: env,
		Release:     rel,
		Debug:       false,
	})
	if err != nil {
		panic(err)
	}
}

func (m Monitor) captureWithFields(capture func(client *sentry.Client, scope sentry.EventModifier)) {
	if !doIngest {
		return
	}

	fields := normalize(m.Fields)
	for k, v := range fields {
		if v == "" {
			fields[k] = "none"
		}
	}
	fields["component"] = comp

	scope := sentry.NewScope()
	scope.SetTags(fields)
	capture(sentryClient, scope)
}
