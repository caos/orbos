package watcher

type Func func(fire chan<- struct{}) error

func (w Func) Watch(fire chan<- struct{}) error {
	return w(fire)
}
