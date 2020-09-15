package zitadel

import (
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/zitadel/configuration"
	"github.com/caos/orbos/internal/secret"
)

func getSecretsMap(desiredKind *DesiredV0) map[string]*secret.Secret {
	secrets := map[string]*secret.Secret{}

	if desiredKind.Spec == nil {
		desiredKind.Spec = &Spec{}
	}

	if desiredKind.Spec.Configuration == nil {
		desiredKind.Spec.Configuration = &configuration.Configuration{}
	}
	conf := desiredKind.Spec.Configuration

	if conf.Tracing == nil {
		conf.Tracing = &configuration.Tracing{}
	}
	if conf.Tracing.ServiceAccountJSON == nil {
		conf.Tracing.ServiceAccountJSON = &secret.Secret{}
	}
	secrets["tracingserviceaccountjson"] = conf.Tracing.ServiceAccountJSON

	if conf.Secrets == nil {
		conf.Secrets = &configuration.Secrets{}
	}
	if conf.Secrets.Keys == nil {
		conf.Secrets.Keys = &secret.Secret{}
	}
	secrets["keys"] = conf.Secrets.Keys

	if conf.Notifications == nil {
		conf.Notifications = &configuration.Notifications{}
	}
	if conf.Notifications.GoogleChatURL == nil {
		conf.Notifications.GoogleChatURL = &secret.Secret{}
	}
	secrets["googlechaturl"] = conf.Notifications.GoogleChatURL

	if conf.Notifications.Twilio.SID == nil {
		conf.Notifications.Twilio.SID = &secret.Secret{}
	}
	secrets["twiliosid"] = conf.Notifications.Twilio.SID

	if conf.Notifications.Twilio.AuthToken == nil {
		conf.Notifications.Twilio.AuthToken = &secret.Secret{}
	}
	secrets["twilioauthtoken"] = conf.Notifications.Twilio.AuthToken

	if conf.Notifications.Email.AppKey == nil {
		conf.Notifications.Email.AppKey = &secret.Secret{}
	}
	secrets["emailappkey"] = conf.Notifications.Email.AppKey

	return secrets
}
