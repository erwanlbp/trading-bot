package eventbus

type Subscription struct {
	EventsCh         chan Event
	EventsSubscribed map[Event]bool
	closed           bool
}

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
func (s *Subscription) Handler(handler func(Event)) {
	for ev := range s.EventsCh {
		handler(ev)
	}
}
