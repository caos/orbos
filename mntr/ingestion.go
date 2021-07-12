package mntr

import (
	"errors"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

// need to be global variables so that clients can be removed too
var (
	sentryClient        *sentry.Client
	env, dsn, comp, rel string
)

func SetContext(version, commit, caosDsn, component, environment string) {
	if rel != "" || dsn != "" {
		panic("SetContext was already called")
	}
	if version == "" || commit == "" {
		panic("version, commit and dsn must not be empty")
	}
	rel = fmt.Sprintf("%s-%s", version, commit)
	dsn = caosDsn
	comp = component
	env = environment

	ingest()
}

func SwitchEnvironment(environment string) {
	if environment == "" {
		panic("environment must not be empty")
	}
	env = environment

	ingest()
}

func ingest() {

	if sentryClient != nil {
		sentryClient.Flush(time.Second * 2)
	}

	if comp == "" || env == "" || rel == "" {
		panic(errors.New("call mntr.SetContext first"))
	}

	var err error
	sentryClient, err = sentry.NewClient(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: fmt.Sprintf("%s-%s", comp, env),
		Release:     rel,
		Debug:       false,
	})
	if err != nil {
		panic(err)
	}
}

func (m Monitor) captureWithFields(capture func(client *sentry.Client, scope sentry.EventModifier)) {
	if sentryClient == nil || sentryClient.Options().Dsn == "" {
		return
	}
	scope := sentry.NewScope()
	scope.SetTags(normalize(m.Fields))
	capture(sentryClient, scope)
}
