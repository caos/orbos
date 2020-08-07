package main

func run(orbconfig string) error {
	newOrbctl, err := buildOrbctl(orbconfig)
	if err != nil {
		return err
	}
	return seq(newOrbctl,
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
