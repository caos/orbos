package core

type NetworkingCurrent interface {
	GetIssuerDomain() string
	GetConsoleDomain() string
	GetAPIDomain() string
	GetAccountsDomain() string
}
