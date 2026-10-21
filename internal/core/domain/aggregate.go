package domain

// Root is embedded by every aggregate root to accumulate the domain events
// raised while handling a command (rdt: core.domain.entity.Aggregate). The
// application layer pulls them after a successful commit and hands them to the
// outbox. Composition, not inheritance: an aggregate embeds Root and gains the
// event bookkeeping without a base-class hierarchy.
//
//	type Dispatch struct {
//	    domain.Root
//	    // ... aggregate state ...
//	}
type Root struct {
	events []DomainEvent
}

// RecordEvents appends events raised by a command, in order.
func (r *Root) RecordEvents(events ...DomainEvent) {
	r.events = append(r.events, events...)
}

// Events returns the events recorded so far without clearing them (rdt:
// Aggregate.events). The slice is a snapshot; mutating it does not affect the
// aggregate.
func (r *Root) Events() []DomainEvent {
	out := make([]DomainEvent, len(r.events))
	copy(out, r.events)
	return out
}

// PullEvents returns the recorded events and clears them in one step — the
// usual move once the work has been persisted (rdt: read Aggregate.events then
// clear_events). After this the aggregate has no pending events.
func (r *Root) PullEvents() []DomainEvent {
	out := r.events
	r.events = nil
	return out
}

// ClearEvents discards any pending events (rdt: Aggregate.clear_events).
func (r *Root) ClearEvents() {
	r.events = nil
}
