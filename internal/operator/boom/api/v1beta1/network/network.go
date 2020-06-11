package network

type Network struct {
	//Defined domain used for external access
	Domain string `json:"domain,omitempty" yaml:"domain,omitempty"`
	//Used email for ACME request
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
	//Used authority for ACME request to get a certificate
	AcmeAuthority string `json:"acmeAuthority,omitempty" yaml:"acmeAuthority,omitempty"`
}
