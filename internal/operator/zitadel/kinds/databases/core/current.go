package core

type DatabaseCurrent interface {
	GetURL() string
	GetPort() string
	GetUsers() []string
}
