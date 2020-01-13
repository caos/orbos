package middleware

import "github.com/caos/orbiter/internal/core/operator/nodeagent"

type Installer interface {
	nodeagent.Installer
	original() nodeagent.Installer
}

func Unwrap(w nodeagent.Installer) nodeagent.Installer {
	if u, ok := w.(Installer); ok {
		return Unwrap(u.original())
	}

	return w
}

type wrapped struct {
	o nodeagent.Installer
}

func (l *wrapped) original() nodeagent.Installer {
	return l.o
}

func (l *wrapped) Is(other nodeagent.Installer) bool {
	return Unwrap(l.o).Is(Unwrap(other))
}

func (l *wrapped) String() string {
	return Unwrap(l.o).String()
}

func (l *wrapped) Equals(other nodeagent.Installer) bool {
	return Unwrap(l.o).Equals(Unwrap(other))
}
