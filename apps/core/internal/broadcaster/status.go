package broadcaster

type BroadcasterStatus uint

const (
	StatusPending BroadcasterStatus = iota
	StatusDistributing
	StatusClosed
)
