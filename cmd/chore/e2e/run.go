package main

func run(orbconfig string) error {
	orbctl := curryOrbctlCommand(orbconfig)
	return destroy(orbctl)
}
