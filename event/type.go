package event

type EventType int

const (
	UNKNOWN EventType = iota
	AUTHORIZED
	UNAUTHORIZED
)

var toString = map[EventType]string{
	UNKNOWN:      "UNKNOWN",
	AUTHORIZED:   "AUTHORIZED",
	UNAUTHORIZED: "UNAUTHORIZED",
}

func (et EventType) String() string {
	return toString[et]
}
