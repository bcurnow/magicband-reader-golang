package event

import (
	"fmt"
)

type Event interface {
	fmt.Stringer
	UID() string
	Type() EventType
	SetType(EventType)
}

type event struct {
	uid       string
	eventType EventType
}

func NewEvent(uid string, eventType EventType) Event {
	return &event{uid: uid, eventType: eventType}
}

func (e *event) UID() string {
	return e.uid
}

func (e *event) Type() EventType {
	return e.eventType
}

func (e *event) SetType(eventType EventType) {
	e.eventType = eventType
}

func (e *event) String() string {
	return fmt.Sprintf("&event.event{uid:\"%v\", eventType:%v}", e.uid, e.eventType.String())
}
