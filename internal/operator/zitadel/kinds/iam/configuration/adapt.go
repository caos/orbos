package configuration

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/configmap"
	secret2 "github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/secret"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/core"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	desired *Configuration,
) (
	func(currentDB interface{}) (resources.EnsureFunc, error),
	resources.DestroyFunc,
	error,
) {
	googleServiceAccountJSONPath := "google-serviceaccount-key.json"
	zitadelKeysPath := "zitadel-keys.yaml"

	tls := ""
	if desired.Email.TLS {
		tls = "TRUE"
	} else {
		tls = "FALSE"
	}

	literalsConfig := map[string]string{
		"GOOGLE_APPLICATION_CREDENTIALS":    "/secret/" + googleServiceAccountJSONPath,
		"ZITADEL_KEY_PATH":                  "/secret/" + zitadelKeysPath,
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
		"TWILIO_SENDER_NAME":                desired.TwilioSenderName,
		"SMTP_HOST":                         desired.Email.SMTPHost,
		"SMTP_USER":                         desired.Email.SMTPUser,
		"EMAIL_SENDER_ADDRESS":              desired.Email.SenderAddress,
		"EMAIL_SENDER_NAME":                 desired.Email.SenderName,
		"SMTP_TLS":                          tls,
		"ZITADEL_ISSUER":                    desired.Endpoints.Issuer,
		"ZITADEL_ACCOUNTS":                  desired.Endpoints.Accounts,
		"ZITADEL_ACCOUNTS_DOMAIN":           desired.Domains.Accounts,
		"ZITADEL_OAUTH":                     desired.Endpoints.OAuth,
		"ZITADEL_AUTHORIZE":                 desired.Endpoints.Authorize,
		"ZITADEL_CONSOLE":                   desired.Endpoints.Console,
		"CAOS_OIDC_DEV":                     "true",
		"ZITADEL_COOKIE_DOMAIN":             desired.Domains.Cookie,
		"ZITADEL_DEFAULT_DOMAIN":            desired.Domains.Default,
		"ZITADEL_CACHE_MAXAGE":              desired.Cache.MaxAge,
		"ZITADEL_CACHE_SHARED_MAXAGE":       desired.Cache.SharedMaxAge,
		"ZITADEL_SHORT_CACHE_MAXAGE":        desired.Cache.ShortMaxAge,
		"ZITADEL_SHORT_CACHE_SHARED_MAXAGE": desired.Cache.ShortSharedMaxAge,
	}

	literalsSecret := map[string]string{}
	if desired.Secrets != nil {
		if desired.Secrets.ServiceAccountJSON != nil {
			literalsSecret[googleServiceAccountJSONPath] = desired.Secrets.ServiceAccountJSON.Value
		}
		if desired.Secrets.Keys != nil {
			literalsSecret[zitadelKeysPath] = desired.Secrets.Keys.Value
		}
	}

	literalsConsoleCM := map[string]string{}
	if desired.ConsoleEnvironmentJSON != nil {
		literalsConsoleCM["environment.json"] = desired.ConsoleEnvironmentJSON.Value
	}

	literalsSecretVars := map[string]string{}
	if desired.SecretVars != nil {
		if desired.SecretVars.EmailAppKey != nil {
			literalsSecretVars["ZITADEL_EMAILAPPKEY"] = desired.SecretVars.EmailAppKey.Value
		}
		if desired.SecretVars.GoogleChatURL != nil {
			literalsSecretVars["ZITADEL_GOOGLE_CHAT_URL"] = desired.SecretVars.GoogleChatURL.Value
		}
		if desired.SecretVars.TwilioAuthToken != nil {
			literalsSecretVars["ZITADEL_TWILIO_AUTH_TOKEN"] = desired.SecretVars.TwilioAuthToken.Value
		}
		if desired.SecretVars.TwilioSID != nil {
			literalsSecretVars["ZITADEL_TWILIO_SID"] = desired.SecretVars.TwilioSID.Value
		}
	}

	_, destroyCM, err := configmap.AdaptFunc("zitadel-vars", namespace, labels, literalsConfig)
	if err != nil {
		return nil, nil, err
	}
	queryS, destroyS, err := secret2.AdaptFunc("zitadel-secret", namespace, labels, literalsSecret)
	if err != nil {
		return nil, nil, err
	}
	queryCCM, destroyCCM, err := configmap.AdaptFunc("console-config", namespace, labels, literalsConsoleCM)
	if err != nil {
		return nil, nil, err
	}
	querySV, destroySV, err := secret2.AdaptFunc("zitadel-secrets-vars", namespace, labels, literalsSecretVars)
	if err != nil {
		return nil, nil, err
	}

	return func(currentDB interface{}) (resources.EnsureFunc, error) {
			current := currentDB.(core.DatabaseCurrent)
			literalsConfig["ZITADEL_EVENTSTORE_HOST"] = current.GetURL()
			literalsConfig["ZITADEL_EVENTSTORE_PORT"] = current.GetPort()
			queryCM, _, err := configmap.AdaptFunc("zitadel-vars", namespace, labels, literalsConfig)
			if err != nil {
				return nil, err
			}
			ensureS, err := queryS()
			if err != nil {
				return nil, err
			}
			ensureCM, err := queryCM()
			if err != nil {
				return nil, err
			}
			ensureCCM, err := queryCCM()
			if err != nil {
				return nil, err
			}
			ensureSV, err := querySV()
			if err != nil {
				return nil, err
			}

			return func(k8sClient *kubernetes.Client) error {
				if err := ensureS(k8sClient); err != nil {
					return err
				}
				if err := ensureCM(k8sClient); err != nil {
					return err
				}
				if err := ensureCCM(k8sClient); err != nil {
					return err
				}
				if err := ensureSV(k8sClient); err != nil {
					return err
				}
				return nil
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			if err := destroyS(k8sClient); err != nil {
				return err
			}
			if err := destroyCM(k8sClient); err != nil {
				return err
			}
			if err := destroyCCM(k8sClient); err != nil {
				return err
			}
			if err := destroySV(k8sClient); err != nil {
				return err
			}
			return nil
		},
		nil
}
