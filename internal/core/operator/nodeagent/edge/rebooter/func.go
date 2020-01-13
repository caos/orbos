package rebooter

type Func func() error

func (f Func) Reboot() error {
	return f()
}
