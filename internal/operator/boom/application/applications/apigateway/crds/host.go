package crds

type HostConfig struct {
	Name             string
	Namespace        string
	Email            string
	AcmeProvider     string
	TLSSecret        string
	PrivateKeySecret string
	Hostname         string
	InsecureAction   string
}
type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}
type PrivateKeySecret struct {
	Name string `yaml:"name,omitempty"`
}
type AcmeProvider struct {
	Authority        string            `yaml:"authority,omitempty"`
	Email            string            `yaml:"email,omitempty"`
	PrivateKeySecret *PrivateKeySecret `yaml:"privateKeySecret,omitempty"`
}

type TLSSecret struct {
	Name string `yaml:"name,omitempty"`
}
type HostSpec struct {
	Hostname     string        `yaml:"hostname"`
	AcmeProvider *AcmeProvider `yaml:"acmeProvider,omitempty"`
	TLSSecret    *TLSSecret    `yaml:"tlsSecret,omitempty"`
}
type Insecure struct {
	Action         string `yaml:"action,omitempty"`
	AdditionalPort string `yaml:"additionalPort,omitempty"`
}

type RequestPolicy struct {
	Insecure *Insecure `yaml:"insecure,omitempty"`
}

type Host struct {
	APIVersion    string         `yaml:"apiVersion"`
	Kind          string         `yaml:"kind"`
	Metadata      *Metadata      `yaml:"metadata"`
	Spec          *HostSpec      `yaml:"spec"`
	RequestPolicy *RequestPolicy `yaml:"requestPolicy,omitempty"`
}

func GetHostFromConfig(conf *HostConfig) *Host {

	var metadata *Metadata
	if conf.Namespace != "" {
		metadata = &Metadata{
			Name:      conf.Name,
			Namespace: conf.Namespace,
		}
	} else {
		metadata = &Metadata{
			Name: conf.Name,
		}
	}

	var tlsSecret *TLSSecret
	if conf.TLSSecret != "" {
		tlsSecret = &TLSSecret{
			Name: conf.TLSSecret,
		}
	}

	var acmeProvider *AcmeProvider
	if conf.AcmeProvider != "" {
		acmeProvider = &AcmeProvider{
			Authority: conf.AcmeProvider,
			Email:     conf.Email,
			// PrivateKeySecret: &PrivateKeySecret{
			// 	Name: conf.PrivateKeySecret,
			// },
		}
	}

	var requestPolicy *RequestPolicy
	if conf.InsecureAction != "" {
		requestPolicy = &RequestPolicy{
			Insecure: &Insecure{
				Action: conf.InsecureAction,
			},
		}
	}

	return &Host{
		APIVersion: "getambassador.io/v2",
		Kind:       "Host",
		Metadata:   metadata,
		Spec: &HostSpec{
			Hostname:     conf.Hostname,
			AcmeProvider: acmeProvider,
			TLSSecret:    tlsSecret,
		},
		RequestPolicy: requestPolicy,
	}
}
