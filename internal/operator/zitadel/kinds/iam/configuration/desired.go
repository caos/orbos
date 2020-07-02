package configuration

import "github.com/caos/orbos/internal/secret"

type Configuration struct {
	Tracing                *Tracing
	TwilioSenderName       string
	Email                  *Email
	Cache                  *Cache
	Domains                *Domains
	Endpoints              *Endpoints
	Secrets                *Secrets
	SecretVars             *SecretVars
	ConsoleEnvironmentJSON *secret.Secret
}

type Secrets struct {
	ServiceAccountJSON *secret.Secret
	Keys               *secret.Secret
	UserVerificationID string
	OTPVerificationID  string
	OIDCKeysID         string
	CookieID           string
	CSRFID             string
}
type SecretVars struct {
	GoogleChatURL   *secret.Secret
	TwilioAuthToken *secret.Secret
	TwilioSID       *secret.Secret
	EmailAppKey     *secret.Secret
}

type Tracing struct {
	ProjectID string
	Fraction  string
}

type Email struct {
	SMTPHost      string
	SMTPUser      string
	SenderAddress string
	SenderName    string
	TLS           bool
}

type Cache struct {
	MaxAge            string
	SharedMaxAge      string
	ShortMaxAge       string
	ShortSharedMaxAge string
}

type Domains struct {
	Accounts string
	Cookie   string
	Default  string
}

type Endpoints struct {
	Authorize string
	OAuth     string
	Issuer    string
	Console   string
	Accounts  string
}
