package configuration

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/configmap"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/secret"
	"github.com/caos/orbos/internal/operator/zitadel"
	coredb "github.com/caos/orbos/internal/operator/zitadel/kinds/databases/core"
	corenw "github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"strings"
)

type ConsoleEnv struct {
	AuthServiceURL string `json:"authServiceUrl"`
	MgmtServiceURL string `json:"mgmtServiceUrl"`
	Issuer         string `json:"issuer"`
	ClientID       string `json:"clientid"`
}

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
	desired *Configuration,
	cmName string,
	certPath string,
	secretName string,
	secretPath string,
	consoleCMName string,
	secretVarsName string,
	secretPasswordName string,
	users map[string]string,
	getClientID func() string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	zitadel.EnsureFunc,
	func(currentDB coredb.DatabaseCurrent, currentNW corenw.NetworkingCurrent) map[string]string,
	error,
) {
	internalMonitor := monitor.WithField("component", "configuration")

	googleServiceAccountJSONPath := "google-serviceaccount-key.json"
	zitadelKeysPath := "zitadel-keys.yaml"

	literalsSecret := literalsSecret(desired, googleServiceAccountJSONPath, zitadelKeysPath)
	literalsSecretVars := literalsSecretVars(desired)

	destroyCM, err := configmap.AdaptFuncToDestroy(namespace, cmName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	destroyS, err := secret.AdaptFuncToDestroy(namespace, secretName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	destroyCCM, err := configmap.AdaptFuncToDestroy(namespace, consoleCMName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	destroySV, err := secret.AdaptFuncToDestroy(namespace, secretVarsName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	destroySP, err := secret.AdaptFuncToDestroy(namespace, secretPasswordName)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyS),
		zitadel.ResourceDestroyToZitadelDestroy(destroyCM),
		zitadel.ResourceDestroyToZitadelDestroy(destroyCCM),
		zitadel.ResourceDestroyToZitadelDestroy(destroySV),
		zitadel.ResourceDestroyToZitadelDestroy(destroySP),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {

			queryS, err := secret.AdaptFuncToEnsure(namespace, secretName, labels, literalsSecret)
			if err != nil {
				return nil, err
			}
			querySV, err := secret.AdaptFuncToEnsure(namespace, secretVarsName, labels, literalsSecretVars)
			if err != nil {
				return nil, err
			}
			querySP, err := secret.AdaptFuncToEnsure(namespace, secretPasswordName, labels, users)
			if err != nil {
				return nil, err
			}

			currentDB, err := coredb.ParseQueriedForDatabase(queried)
			if err != nil {
				return nil, err
			}

			currentNW, err := corenw.ParseQueriedForNetworking(queried)
			if err != nil {
				return nil, err
			}

			queryCCM, err := configmap.AdaptFuncToEnsure(namespace, consoleCMName, labels, literalsConsoleCM(getClientID(), currentNW))
			if err != nil {
				return nil, err
			}

			queryCM, err := configmap.AdaptFuncToEnsure(namespace, cmName, labels, literalsConfigMap(desired, users, certPath, secretPath, googleServiceAccountJSONPath, zitadelKeysPath, currentNW, currentDB))
			if err != nil {
				return nil, err
			}

			queriers := []zitadel.QueryFunc{
				zitadel.ResourceQueryToZitadelQuery(queryS),
				zitadel.ResourceQueryToZitadelQuery(queryCCM),
				zitadel.ResourceQueryToZitadelQuery(querySV),
				zitadel.ResourceQueryToZitadelQuery(querySP),
				zitadel.ResourceQueryToZitadelQuery(queryCM),
			}

			return zitadel.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		func(k8sClient *kubernetes.Client) error {
			monitor.Debug("Waiting for configuration to be created")
			if err := k8sClient.WaitForSecret(namespace, secretName, 60); err != nil {
				return errors.Wrap(err, "error while waiting for secret")
			}

			if err := k8sClient.WaitForSecret(namespace, secretVarsName, 60); err != nil {
				return errors.Wrap(err, "error while waiting for vars secret ")
			}

			if err := k8sClient.WaitForSecret(namespace, secretPasswordName, 60); err != nil {
				return errors.Wrap(err, "error while waiting for password secret")
			}

			if err := k8sClient.WaitForConfigMap(namespace, cmName, 60); err != nil {
				return errors.Wrap(err, "error while waiting for configmap")
			}

			if err := k8sClient.WaitForConfigMap(namespace, consoleCMName, 60); err != nil {
				return errors.Wrap(err, "error while waiting for console configmap")
			}
			monitor.Debug("configuration is created")
			return nil
		},
		func(currentDB coredb.DatabaseCurrent, currentNW corenw.NetworkingCurrent) map[string]string {
			return map[string]string{
				secretName:         getHash(literalsSecret),
				secretVarsName:     getHash(literalsSecretVars),
				secretPasswordName: getHash(users),
				cmName:             getHash(literalsConfigMap(desired, users, certPath, secretPath, googleServiceAccountJSONPath, zitadelKeysPath, currentNW, currentDB)),
				consoleCMName:      getHash(literalsConsoleCM(getClientID(), currentNW)),
			}
		},
		nil
}

func getHash(dataMap map[string]string) string {
	data, err := json.Marshal(dataMap)
	if err != nil {
		return ""
	}
	h := sha512.New()
	return base64.URLEncoding.EncodeToString(h.Sum(data))
}

func literalsConfigMap(
	desired *Configuration,
	users map[string]string,
	certPath, secretPath, googleServiceAccountJSONPath, zitadelKeysPath string,
	currentNW corenw.NetworkingCurrent,
	currentDB coredb.DatabaseCurrent,
) map[string]string {

	tls := ""
	if desired.Notifications.Email.TLS {
		tls = "TRUE"
	} else {
		tls = "FALSE"
	}

	literalsConfigMap := map[string]string{
		"GOOGLE_APPLICATION_CREDENTIALS": secretPath + "/" + googleServiceAccountJSONPath,
		"ZITADEL_KEY_PATH":               secretPath + "/" + zitadelKeysPath,
		"ZITADEL_LOG_LEVEL":              "debug",
		"DEBUG_MODE":                     "TRUE",
		"SMTP_TLS":                       tls,
		"CAOS_OIDC_DEV":                  "true",
		"CR_SSL_MODE":                    "require",
		"CR_ROOT_CERT":                   certPath + "/ca.crt",
	}

	if users != nil {
		for _, user := range users {
			literalsConfigMap["CR_"+strings.ToUpper(user)+"_CERT"] = certPath + "/client." + user + ".crt"
			literalsConfigMap["CR_"+strings.ToUpper(user)+"_KEY"] = certPath + "/client." + user + ".key"
		}
	}

	if desired != nil {
		if desired.Tracing != nil {
			literalsConfigMap["ZITADEL_TRACING_PROJECT_ID"] = desired.Tracing.ProjectID
			literalsConfigMap["ZITADEL_TRACING_FRACTION"] = desired.Tracing.Fraction
		}
		if desired.Secrets != nil {
			literalsConfigMap["ZITADEL_USER_VERIFICATION_KEY"] = desired.Secrets.UserVerificationID
			literalsConfigMap["ZITADEL_OTP_VERIFICATION_KEY"] = desired.Secrets.OTPVerificationID
			literalsConfigMap["ZITADEL_OIDC_KEYS_ID"] = desired.Secrets.OIDCKeysID
			literalsConfigMap["ZITADEL_COOKIE_KEY"] = desired.Secrets.CookieID
			literalsConfigMap["ZITADEL_CSRF_KEY"] = desired.Secrets.CSRFID
		}
		if desired.Notifications != nil {
			literalsConfigMap["TWILIO_SENDER_NAME"] = desired.Notifications.Twilio.SenderName
			literalsConfigMap["SMTP_HOST"] = desired.Notifications.Email.SMTPHost
			literalsConfigMap["SMTP_USER"] = desired.Notifications.Email.SMTPUser
			literalsConfigMap["EMAIL_SENDER_ADDRESS"] = desired.Notifications.Email.SenderAddress
			literalsConfigMap["EMAIL_SENDER_NAME"] = desired.Notifications.Email.SenderName
		}
		if desired.Cache != nil {
			literalsConfigMap["ZITADEL_CACHE_MAXAGE"] = desired.Cache.MaxAge
			literalsConfigMap["ZITADEL_CACHE_SHARED_MAXAGE"] = desired.Cache.SharedMaxAge
			literalsConfigMap["ZITADEL_SHORT_CACHE_MAXAGE"] = desired.Cache.ShortMaxAge
			literalsConfigMap["ZITADEL_SHORT_CACHE_SHARED_MAXAGE"] = desired.Cache.ShortSharedMaxAge
		}
	}

	if currentDB != nil {
		literalsConfigMap["ZITADEL_EVENTSTORE_HOST"] = currentDB.GetURL()
		literalsConfigMap["ZITADEL_EVENTSTORE_PORT"] = currentDB.GetPort()
	}

	if currentNW != nil {
		defaultDomain := currentNW.GetDomain()
		accountsDomain := currentNW.GetAccountsSubDomain() + "." + defaultDomain
		accounts := "https://" + accountsDomain
		issuer := "https://" + currentNW.GetIssuerSubDomain() + "." + defaultDomain
		oauth := "https://" + currentNW.GetAPISubDomain() + "." + defaultDomain + "/oauth/v2"
		authorize := "https://" + currentNW.GetAccountsSubDomain() + "." + defaultDomain + "/oauth/v2"
		console := "https://" + currentNW.GetConsoleSubDomain() + "." + defaultDomain

		literalsConfigMap["ZITADEL_ISSUER"] = issuer
		literalsConfigMap["ZITADEL_ACCOUNTS"] = accounts
		literalsConfigMap["ZITADEL_OAUTH"] = oauth
		literalsConfigMap["ZITADEL_AUTHORIZE"] = authorize
		literalsConfigMap["ZITADEL_CONSOLE"] = console
		literalsConfigMap["ZITADEL_ACCOUNTS_DOMAIN"] = accountsDomain
		literalsConfigMap["ZITADEL_COOKIE_DOMAIN"] = accountsDomain
		literalsConfigMap["ZITADEL_DEFAULT_DOMAIN"] = defaultDomain
	}

	return literalsConfigMap
}

func literalsSecret(desired *Configuration, googleServiceAccountJSONPath, zitadelKeysPath string) map[string]string {
	literalsSecret := map[string]string{}
	if desired != nil {
		if desired.Tracing != nil && desired.Tracing.ServiceAccountJSON != nil {
			literalsSecret[googleServiceAccountJSONPath] = desired.Tracing.ServiceAccountJSON.Value
		}
		if desired.Secrets != nil && desired.Secrets.Keys != nil {
			literalsSecret[zitadelKeysPath] = desired.Secrets.Keys.Value
		}
	}
	return literalsSecret
}

func literalsSecretVars(desired *Configuration) map[string]string {
	literalsSecretVars := map[string]string{}
	if desired != nil {
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
	}
	return literalsSecretVars
}

func literalsConsoleCM(
	clientID string,
	currentNW corenw.NetworkingCurrent,
) map[string]string {
	literalsConsoleCM := map[string]string{}
	consoleEnv := ConsoleEnv{
		ClientID: clientID,
	}

	if currentNW != nil {
		defaultDomain := currentNW.GetDomain()
		consoleEnv.Issuer = "https://" + currentNW.GetIssuerSubDomain() + "." + defaultDomain
		consoleEnv.AuthServiceURL = "https://" + currentNW.GetAPISubDomain() + "." + defaultDomain
		consoleEnv.MgmtServiceURL = "https://" + currentNW.GetAPISubDomain() + "." + defaultDomain
	}
	data, err := json.Marshal(consoleEnv)
	if err != nil {
		return map[string]string{}
	}

	literalsConsoleCM["environment.json"] = string(data)
	return literalsConsoleCM
}
