package middleware

import "github.com/caos/orbiter/internal/kinds/nodeagent/adapter"

type Installer interface {
	adapter.Installer
	original() adapter.Installer
}

func Unwrap(w adapter.Installer) adapter.Installer {
	if u, ok := w.(Installer); ok {
		return Unwrap(u.original())
	}

	return w
}

type wrapped struct {
	o adapter.Installer
}

func (l *wrapped) original() adapter.Installer {
	return l.o
}

func (l *wrapped) Is(other adapter.Installer) bool {
	return Unwrap(l.o).Is(Unwrap(other))
}

func (l *wrapped) String() string {
	return Unwrap(l.o).String()
}

func (l *wrapped) Equals(other adapter.Installer) bool {
	return Unwrap(l.o).Equals(Unwrap(other))
}
