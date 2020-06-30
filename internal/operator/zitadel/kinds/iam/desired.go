package iam

import (
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type DesiredV0 struct {
	Common   *tree.Common `yaml:",inline"`
	Spec     Spec
	Database *tree.Tree
}

type Spec struct {
	Verbose          bool
	ReplicaCount     int `yaml:"replicaCount,omitempty"`
	Tracing          *Tracing
	TwilioSenderName string
	Email            *Email
	Cache            *Cache
	Domains          *Domains
	Endpoints        *Endpoints
	Secrets          *Secrets
}

type Secrets struct {
	ServiceAccountJSON *secret.Secret
	UserVerification   *secret.Secret
	OTPVerification    *secret.Secret
	OIDCKeysID         *secret.Secret
	Cookie             *secret.Secret
	CSRF               *secret.Secret
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
}

type Endpoints struct {
	Authorize string
	OAuth     string
	Issuer    string
	Console   string
	Accounts  string
}

func parseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{
		Common: desiredTree.Common,
		Spec:   Spec{},
	}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
