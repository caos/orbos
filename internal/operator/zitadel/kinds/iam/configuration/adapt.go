package configuration

import (
	"errors"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/configmap"
	secret2 "github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/secret"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/core"
	"github.com/caos/orbos/internal/tree"
	"strings"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	desired *Configuration,
	cmName string,
	certPath string,
	secretName string,
	secretPath string,
	consoleCMName string,
	secretVarsName string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	googleServiceAccountJSONPath := "google-serviceaccount-key.json"
	zitadelKeysPath := "zitadel-keys.yaml"

	tls := ""
	if desired.Notifications.Email.TLS {
		tls = "TRUE"
	} else {
		tls = "FALSE"
	}

	literalsConfig := map[string]string{
		"GOOGLE_APPLICATION_CREDENTIALS":    secretPath + "/" + googleServiceAccountJSONPath,
		"ZITADEL_KEY_PATH":                  secretPath + "/" + zitadelKeysPath,
		"ZITADEL_TRACING_PROJECT_ID":        desired.Tracing.ProjectID,
		"ZITADEL_TRACING_FRACTION":          desired.Tracing.Fraction,
		"ZITADEL_LOG_LEVEL":                 "debug",
		"ZITADEL_EVENTSTORE_HOST":           "",
		"ZITADEL_EVENTSTORE_PORT":           "",
		"ZITADEL_USER_VERIFICATION_KEY":     desired.Secrets.UserVerificationID,
		"ZITADEL_OTP_VERIFICATION_KEY":      desired.Secrets.OTPVerificationID,
		"ZITADEL_OIDC_KEYS_ID":              desired.Secrets.OIDCKeysID,
		"ZITADEL_COOKIE_KEY":                desired.Secrets.CookieID,
		"ZITADEL_CSRF_KEY":                  desired.Secrets.CSRFID,
		"DEBUG_MODE":                        "TRUE",
		"TWILIO_SENDER_NAME":                desired.Notifications.Twilio.SenderName,
		"SMTP_HOST":                         desired.Notifications.Email.SMTPHost,
		"SMTP_USER":                         desired.Notifications.Email.SMTPUser,
		"EMAIL_SENDER_ADDRESS":              desired.Notifications.Email.SenderAddress,
		"EMAIL_SENDER_NAME":                 desired.Notifications.Email.SenderName,
		"SMTP_TLS":                          tls,
		"ZITADEL_ISSUER":                    desired.Endpoints.Issuer,
		"ZITADEL_ACCOUNTS":                  desired.Endpoints.Accounts,
		"ZITADEL_OAUTH":                     desired.Endpoints.OAuth,
		"ZITADEL_AUTHORIZE":                 desired.Endpoints.Authorize,
		"ZITADEL_CONSOLE":                   desired.Endpoints.Console,
		"ZITADEL_ACCOUNTS_DOMAIN":           desired.Domains.Accounts,
		"ZITADEL_COOKIE_DOMAIN":             desired.Domains.Cookie,
		"ZITADEL_DEFAULT_DOMAIN":            desired.Domains.Default,
		"CAOS_OIDC_DEV":                     "true",
		"ZITADEL_CACHE_MAXAGE":              desired.Cache.MaxAge,
		"ZITADEL_CACHE_SHARED_MAXAGE":       desired.Cache.SharedMaxAge,
		"ZITADEL_SHORT_CACHE_MAXAGE":        desired.Cache.ShortMaxAge,
		"ZITADEL_SHORT_CACHE_SHARED_MAXAGE": desired.Cache.ShortSharedMaxAge,
		"CR_SSL_MODE":                       "require",
		"CR_ROOT_CERT":                      certPath + "/ca.crt",
	}

	userList := []string{"management", "auth", "authz", "adminapi", "notification"}
	for _, user := range userList {
		literalsConfig["CR_"+strings.ToUpper(user)+"_CERT"] = certPath + "/client." + user + ".crt"
		literalsConfig["CR_"+strings.ToUpper(user)+"_KEY"] = certPath + "/client." + user + ".key"
		literalsConfig["CR_"+strings.ToUpper(user)+"_PASSWORD"] = user
	}

	literalsSecret := map[string]string{}

	if desired.Tracing != nil && desired.Tracing.ServiceAccountJSON != nil {
		literalsSecret[googleServiceAccountJSONPath] = desired.Tracing.ServiceAccountJSON.Value
	}
	if desired.Secrets != nil && desired.Secrets.Keys != nil {
		literalsSecret[zitadelKeysPath] = desired.Secrets.Keys.Value
	}

	literalsConsoleCM := map[string]string{}
	if desired.ConsoleEnvironmentJSON != nil {
		literalsConsoleCM["environment.json"] = desired.ConsoleEnvironmentJSON.Value
	}

	literalsSecretVars := map[string]string{}
	if desired.Notifications != nil {
		if desired.Notifications.Email.AppKey != nil {
			literalsSecretVars["ZITADEL_EMAILAPPKEY"] = desired.Notifications.Email.AppKey.Value
		}
		if desired.Notifications.GoogleChatURL != nil {
			literalsSecretVars["ZITADEL_GOOGLE_CHAT_URL"] = desired.Notifications.GoogleChatURL.Value
		}
		if desired.Notifications.Twilio.AuthToken != nil {
			literalsSecretVars["ZITADEL_TWILIO_AUTH_TOKEN"] = desired.Notifications.Twilio.AuthToken.Value
		}
		if desired.Notifications.Twilio.SID != nil {
			literalsSecretVars["ZITADEL_TWILIO_SID"] = desired.Notifications.Twilio.SID.Value
		}
	}

	_, destroyCM, err := configmap.AdaptFunc(cmName, namespace, labels, literalsConfig)
	if err != nil {
		return nil, nil, err
	}
	queryS, destroyS, err := secret2.AdaptFunc(secretName, namespace, labels, literalsSecret)
	if err != nil {
		return nil, nil, err
	}
	queryCCM, destroyCCM, err := configmap.AdaptFunc(consoleCMName, namespace, labels, literalsConsoleCM)
	if err != nil {
		return nil, nil, err
	}
	querySV, destroySV, err := secret2.AdaptFunc(secretVarsName, namespace, labels, literalsSecretVars)
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{
		zitadel.ResourceQueryToZitadelQuery(queryS),
		zitadel.ResourceQueryToZitadelQuery(queryCCM),
		zitadel.ResourceQueryToZitadelQuery(querySV),
	}

	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyS),
		zitadel.ResourceDestroyToZitadelDestroy(destroyCM),
		zitadel.ResourceDestroyToZitadelDestroy(destroyCCM),
		zitadel.ResourceDestroyToZitadelDestroy(destroySV),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			queriedDB, ok := queried["database"]
			if !ok {
				return nil, errors.New("no current state for database found")
			}
			current, ok := queriedDB.(*tree.Tree)
			if !ok {
				return nil, errors.New("current state does not fullfil interface")
			}
			currentDB, ok := current.Parsed.(core.DatabaseCurrent)
			if !ok {
				return nil, errors.New("current state does not fullfil interface")
			}

			literalsConfig["ZITADEL_EVENTSTORE_HOST"] = currentDB.GetURL()
			literalsConfig["ZITADEL_EVENTSTORE_PORT"] = currentDB.GetPort()
			queryCM, _, err := configmap.AdaptFunc(cmName, namespace, labels, literalsConfig)
			if err != nil {
				return nil, err
			}

			queriers = append(queriers, zitadel.ResourceQueryToZitadelQuery(queryCM))

			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}
