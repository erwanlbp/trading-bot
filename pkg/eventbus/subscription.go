package eventbus

import (
	"context"
)

type Subscription struct {
	EventsCh         chan Event
	EventsSubscribed map[string]bool
	closed           bool
}

type EventHandler func(context.Context, Event)

func newSubscription(events []string) *Subscription {
	eventsMap := make(map[string]bool, len(events))
	for _, event := range events {
		eventsMap[event] = true
	}
	return &Subscription{
		EventsCh:         make(chan Event),
		EventsSubscribed: eventsMap,
	}
}

func (s *Subscription) IsSubscribed(event string) bool {
	_, ok := s.EventsSubscribed[event]
	return ok
}

func (s *Subscription) Close() {
	if s.closed {
		return
	}
	s.closed = true
	close(s.EventsCh)
}

// If events is not provided, will be all events for this subscription
func (s *Subscription) Handler(ctx context.Context, handler EventHandler) {
	go func() {
		<-ctx.Done()
		s.Close()
	}()
	for ev := range s.EventsCh {
		handler(ctx, ev)
	}
}
