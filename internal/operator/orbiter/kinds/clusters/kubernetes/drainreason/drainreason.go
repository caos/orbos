package drainreason

type DrainReason int

const (
	Updating DrainReason = iota
	Rebooting
	Deleting
)
