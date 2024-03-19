package eventbus

import "context"

type Subscription struct {
	EventsCh         chan Event
	EventsSubscribed map[Event]bool
	closed           bool
}

type EventHandler func(context.Context, Event)

func newSubscription(events []Event) *Subscription {
	eventsMap := make(map[Event]bool)
	for _, event := range events {
		eventsMap[event] = true
	}
	return &Subscription{
		EventsCh:         make(chan Event),
		EventsSubscribed: eventsMap,
	}
}

func (s *Subscription) Close() {
	s.closed = true
	close(s.EventsCh)
}

// If events is not provided, will be all events for this subscription
func (s *Subscription) Handler(ctx context.Context, handler EventHandler) {
	for {
		select {
		case ev := <-s.EventsCh:
			handler(ctx, ev)
		case <-ctx.Done():
			break
		}
	}
}
