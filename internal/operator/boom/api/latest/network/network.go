package network

type Network struct {
	Domain        string `json:"domain,omitempty" yaml:"domain,omitempty"`
	Email         string `json:"email,omitempty" yaml:"email,omitempty"`
	AcmeAuthority string `json:"acmeAuthority,omitempty" yaml:"acmeAuthority,omitempty"`
}
