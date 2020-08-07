package main

func run(orbconfig string) error {
	return seq(curryOrbctlCommand(orbconfig),
		initORBITER,
		destroy,
		bootstrap,
	)
}

func seq(orbctl newOrbctlCommandFunc, fns ...func(newOrbctlCommandFunc) error) error {
	for _, fn := range fns {
		if err := fn(orbctl); err != nil {
			return err
		}
	}
	return nil
}
