package configuration

import "github.com/caos/orbos/internal/secret"

type Configuration struct {
	//Tracing configuration for zitadel
	Tracing *Tracing `yaml:"tracing,omitempty"`
	//Cache configuration for zitadel
	Cache *Cache `yaml:"cache,omitempty"`
	//Secrets used by zitadel
	Secrets *Secrets `yaml:"secrets,omitempty"`
	//Notification configuration for zitadel
	Notifications *Notifications `yaml:"notifications,omitempty"`
	//Passwords used for the maintaining of the users in the database
	Passwords *Passwords `yaml:"passwords,omitempty"`
	//Debug mode for zitadel if notifications should be only sent by chat
	DebugMode bool `yaml:"debugMode"`
	//Log-level for zitadel
	LogLevel string `yaml:"logLevel"`
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
	//Password for the User "migration"
	Migration *secret.Secret `yaml:"migration"`
	//Password for the User "management"
	Management *secret.Secret `yaml:"management"`
	//Password for the User "auth"
	Auth *secret.Secret `yaml:"auth"`
	//Password for the User "authz"
	Authz *secret.Secret `yaml:"authz"`
	//Password for the User "adminapi"
	Adminapi *secret.Secret `yaml:"adminapi"`
	//Password for the User "notification"
	Notification *secret.Secret `yaml:"notification"`
	//Password for the User "eventstore"
	Eventstore *secret.Secret `yaml:"eventstore"`
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
	//Text-file which consists of a list of key/value to provide the keys to encrypt data in zitadel
	Keys *secret.Secret `yaml:"keys,omitempty"`
	//Key used from keys-file for user verification
	UserVerificationID string `yaml:"userVerificationID,omitempty"`
	//Key used from keys-file for OTP verification
	OTPVerificationID string `yaml:"otpVerificationID,omitempty"`
	//Key used from keys-file for OIDC
	OIDCKeysID string `yaml:"oidcKeysID,omitempty"`
	//Key used from keys-file for cookies
	CookieID string `yaml:"cookieID,omitempty"`
	//Key used from keys-file for CSRF
	CSRFID string `yaml:"csrfID,omitempty"`
	//Key used from keys-file for domain verification
	DomainVerificationID string `yaml:"domainVerificationID,omitempty"`
	//Key used from keys-file for IDP configuration verification
	IDPConfigVerificationID string `yaml:"idpConfigVerificationID,omitempty"`
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
	//Google chat URL used for notifications
	GoogleChatURL *secret.Secret `yaml:"googleChatURL,omitempty"`
	//Configuration for email notifications
	Email *Email `yaml:"email,omitempty"`
	//Configuration for twilio notifications
	Twilio *Twilio `yaml:"twilio,omitempty"`
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
	//Sender name for Twilio
	SenderName string `yaml:"senderName,omitempty"`
	//Auth token to connect with Twilio
	AuthToken *secret.Secret `yaml:"authToken,omitempty"`
	//SID to connect with Twilio
	SID *secret.Secret `yaml:"sid,omitempty"`
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
	//SMTP host used for email notifications
	SMTPHost string `yaml:"smtpHost,omitempty"`
	//SMTP user used for email notifications
	SMTPUser string `yaml:"smtpUser,omitempty"`
	//Sender address from where the emails should get sent
	SenderAddress string `yaml:"senderAddress,omitempty"`
	//Sender name form where the emails should get sent
	SenderName string `yaml:"senderName,omitempty"`
	//Flag if TLS should be used for the communication with the SMTP host
	TLS bool `yaml:"tls,omitempty"`
	//Application-key used for SMTP communication
	AppKey *secret.Secret `yaml:"appKey,omitempty"`
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
	//Max age for cache records
	MaxAge string `yaml:"maxAge,omitempty"`
	//Max age for the shared cache records
	SharedMaxAge string `yaml:"sharedMaxAge,omitempty"`
	//Max age for the short cache records
	ShortMaxAge string `yaml:"shortMaxAge,omitempty"`
	//Max age for the short shared cache records
	ShortSharedMaxAge string `yaml:"shortSharedMaxAge,omitempty"`
}
