package core

type NetworkingCurrent interface {
	GetDomain() string
	GetIssuerSubDomain() string
	GetConsoleSubDomain() string
	GetAPISubDomain() string
	GetAccountsSubDomain() string
}
