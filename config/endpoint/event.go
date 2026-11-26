package endpoint

import (
	"time"
)

// Event is something that happens at a specific time
type Event struct {
	// Type is the kind of event
	Type EventType `json:"type"`

	// Timestamp is the moment at which the event happened
	Timestamp time.Time `json:"timestamp"`
}

// EventType is, uh, the types of events?
type EventType string

var (
	// EventStart is a type of event that represents when an endpoint starts being monitored
	EventStart EventType = "START"

	// EventHealthy is a type of event that represents an endpoint passing all of its conditions
	EventHealthy EventType = "HEALTHY"

	// EventUnhealthy is a type of event that represents an endpoint failing one or more of its conditions
	EventUnhealthy EventType = "UNHEALTHY"

	// EventDegraded is a type of event that represents an endpoint that passes its conditions but fails one or more of
	// its degraded-conditions (slow, for instance). It is not downtime: uptime is not affected.
	EventDegraded EventType = "DEGRADED"
)

// NewEventFromResult creates an Event from a Result
func NewEventFromResult(result *Result) *Event {
	event := &Event{Timestamp: result.Timestamp}
	switch {
	case !result.Success:
		event.Type = EventUnhealthy
	case result.Degraded():
		event.Type = EventDegraded
	default:
		event.Type = EventHealthy
	}
	return event
}
