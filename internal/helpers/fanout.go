package helpers

import "sync"

func Fanout(funcs []func() error) func() error {
	return func() error {
		var wg sync.WaitGroup
		var err error
		for _, f := range funcs {
			wg.Add(1)
			go func(f func() error) {
				err = Concat(err, f())
				wg.Done()
			}(f)
		}
		wg.Wait()
		return err
	}
}
