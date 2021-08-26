package event

import (
	"fmt"
)

type EventType int

const (
	UNKNOWN EventType = iota
	AUTHORIZED
	UNAUTHORIZED
)

type Event interface {
	fmt.Stringer
	UID() string
	Type() EventType
}

type event struct {
	uid       string
	eventType EventType
}

func (e *event) UID() string {
	return e.uid
}

func (e *event) Type() EventType {
	return e.eventType
}

func (e *event) String() string {
	return fmt.Sprintf("%#v", e)
}

func NewEvent(uid string, eventType EventType) *event {
	return &event{uid: uid, eventType: eventType}
}
