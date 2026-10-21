package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent is implemented by every domain event (rdt: core.domain.event.
// DomainEvent). SubjectID is the id of the aggregate the event is about — what
// the transactional outbox records as aggregate_id — returned as a raw UUID
// because the outbox is generic and does not know the concrete id type.
type DomainEvent interface {
	EventID() uuid.UUID
	OccurredAt() time.Time
	SubjectID() uuid.UUID
}

// BaseEvent supplies the event_id and occurred_at that every event carries.
// Embed it in each concrete event struct and construct it with NewBaseEvent;
// the concrete event supplies SubjectID.
//
//	type DispatchQueued struct {
//	    domain.BaseEvent
//	    DispatchID DispatchID
//	}
//	func (e DispatchQueued) SubjectID() uuid.UUID { return e.DispatchID.UUID() }
type BaseEvent struct {
	eventID    uuid.UUID
	occurredAt time.Time
}

// NewBaseEvent stamps a fresh event id (v7, time-ordered) and the current UTC
// instant, matching rdt's default_factory fields.
func NewBaseEvent() BaseEvent {
	return BaseEvent{
		eventID:    uuid.Must(uuid.NewV7()),
		occurredAt: time.Now().UTC(),
	}
}

// EventID returns this event's own identity.
func (e BaseEvent) EventID() uuid.UUID { return e.eventID }

// OccurredAt returns when the event was recorded (UTC).
func (e BaseEvent) OccurredAt() time.Time { return e.occurredAt }
