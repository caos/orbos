package configuration

import "github.com/caos/orbos/internal/secret"

type Configuration struct {
	Tracing       *Tracing       `yaml:"tracing,omitempty"`
	Cache         *Cache         `yaml:"cache,omitempty"`
	Secrets       *Secrets       `yaml:"secrets,omitempty"`
	Notifications *Notifications `yaml:"notifications,omitempty"`
	Passwords     *Passwords     `yaml:"passwords,omitempty"`
	DebugMode     bool           `yaml:"debugMode"`
	LogLevel      string         `yaml:"logLevel"`
}

func (c *Configuration) IsZero() bool {
	if (c.Tracing == nil || c.Tracing.IsZero()) &&
		(c.Cache == nil) &&
		(c.Secrets == nil || c.Secrets.IsZero()) &&
		(c.Notifications == nil || c.Notifications.IsZero()) &&
		(c.Passwords == nil || c.Passwords.IsZero()) &&
		!c.DebugMode &&
		c.LogLevel == "" {
		return true
	}
	return false
}

type Passwords struct {
	Migration    *secret.Secret `yaml:"migration"`
	Management   *secret.Secret `yaml:"management"`
	Auth         *secret.Secret `yaml:"auth"`
	Authz        *secret.Secret `yaml:"authz"`
	Adminapi     *secret.Secret `yaml:"adminapi"`
	Notification *secret.Secret `yaml:"notification"`
	Eventstore   *secret.Secret `yaml:"eventstore"`
}

func (p *Passwords) IsZero() bool {
	if (p.Migration == nil || p.Migration.IsZero()) &&
		(p.Management == nil || p.Management.IsZero()) &&
		(p.Auth == nil || p.Auth.IsZero()) &&
		(p.Authz == nil || p.Authz.IsZero()) &&
		(p.Adminapi == nil || p.Adminapi.IsZero()) &&
		(p.Notification == nil || p.Notification.IsZero()) &&
		(p.Eventstore == nil || p.Eventstore.IsZero()) {
		return true
	}
	return false
}

type Secrets struct {
	Keys                    *secret.Secret `yaml:"keys,omitempty"`
	UserVerificationID      string         `yaml:"userVerificationID,omitempty"`
	OTPVerificationID       string         `yaml:"otpVerificationID,omitempty"`
	OIDCKeysID              string         `yaml:"oidcKeysID,omitempty"`
	CookieID                string         `yaml:"cookieID,omitempty"`
	CSRFID                  string         `yaml:"csrfID,omitempty"`
	DomainVerificationID    string         `yaml:"domainVerificationID,omitempty"`
	IDPConfigVerificationID string         `yaml:"idpConfigVerificationID,omitempty"`
}

func (s *Secrets) IsZero() bool {
	if (s.Keys == nil || s.Keys.IsZero()) &&
		s.UserVerificationID == "" &&
		s.OTPVerificationID == "" &&
		s.OIDCKeysID == "" &&
		s.CookieID == "" &&
		s.CSRFID == "" &&
		s.DomainVerificationID == "" &&
		s.IDPConfigVerificationID == "" {
		return true
	}
	return false
}

type Notifications struct {
	GoogleChatURL *secret.Secret `yaml:"googleChatURL,omitempty"`
	Email         *Email         `yaml:"email,omitempty"`
	Twilio        *Twilio        `yaml:"twilio,omitempty"`
}

func (n *Notifications) IsZero() bool {
	if (n.GoogleChatURL == nil || n.GoogleChatURL.IsZero()) &&
		(n.Email == nil || n.Email.IsZero()) &&
		(n.Twilio == nil || n.Twilio.IsZero()) {
		return true
	}
	return false

}

type Tracing struct {
	ServiceAccountJSON *secret.Secret `yaml:"serviceAccountJSON,omitempty"`
	ProjectID          string         `yaml:"projectID,omitempty"`
	Fraction           string         `yaml:"fraction,omitempty"`
	Type               string         `yaml:"type,omitempty"`
}

func (t *Tracing) IsZero() bool {
	if (t.ServiceAccountJSON == nil || t.ServiceAccountJSON.IsZero()) &&
		t.ProjectID == "" &&
		t.Fraction == "" &&
		t.Type == "" {
		return true
	}
	return false
}

type Twilio struct {
	SenderName string         `yaml:"senderName,omitempty"`
	AuthToken  *secret.Secret `yaml:"authToken,omitempty"`
	SID        *secret.Secret `yaml:"sid,omitempty"`
}

func (t *Twilio) IsZero() bool {
	if (t.SID == nil || t.SID.IsZero()) &&
		(t.AuthToken == nil || t.AuthToken.IsZero()) &&
		t.SenderName == "" {
		return true
	}
	return false
}

type Email struct {
	SMTPHost      string         `yaml:"smtpHost,omitempty"`
	SMTPUser      string         `yaml:"smtpUser,omitempty"`
	SenderAddress string         `yaml:"senderAddress,omitempty"`
	SenderName    string         `yaml:"senderName,omitempty"`
	TLS           bool           `yaml:"tls,omitempty"`
	AppKey        *secret.Secret `yaml:"appKey,omitempty"`
}

func (e *Email) IsZero() bool {
	if (e.AppKey == nil || e.AppKey.IsZero()) &&
		!e.TLS &&
		e.SMTPHost == "" &&
		e.SMTPUser == "" &&
		e.SenderAddress == "" &&
		e.SenderName == "" {
		return true
	}
	return false
}

type Cache struct {
	MaxAge            string `yaml:"maxAge,omitempty"`
	SharedMaxAge      string `yaml:"sharedMaxAge,omitempty"`
	ShortMaxAge       string `yaml:"shortMaxAge,omitempty"`
	ShortSharedMaxAge string `yaml:"shortSharedMaxAge,omitempty"`
}
