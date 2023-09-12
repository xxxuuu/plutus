package app

import (
	"time"
)

type EventContext struct {
	event *Event
	data  map[any]any
}

func NewEventContext(e *Event) EventContext {
	return EventContext{
		event: e,
		data:  map[any]any{},
	}
}

func (c EventContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c EventContext) Done() <-chan struct{} {
	return nil
}

func (c EventContext) Err() error {
	return nil
}

func (c EventContext) Value(key any) any {
	return c.data[key]
}

func (c EventContext) Set(key any, value any) {
	c.data[key] = value
}

func (c EventContext) Event() *Event {
	return c.event
}
