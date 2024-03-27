package eventbus

import "sync"

type Bus struct {
	mtx           sync.RWMutex
	subscriptions []*Subscription
}

func NewEventBus() *Bus {
	return &Bus{}
}

func (b *Bus) Notify(event Event) {
	go func() {
		b.mtx.RLock()
		defer b.mtx.RUnlock()

		for i := 0; i < len(b.subscriptions); i += 1 {
			sub := b.subscriptions[i]

			if _, ok := sub.EventsSubscribed[event.Name]; ok {
				// TODO What if this sub is not listening right now, that'll block this func
				sub.EventsCh <- event
			}
		}
	}()
}

func (b *Bus) Subscribe(events ...string) *Subscription {
	sub := newSubscription(events)

	b.mtx.Lock()
	defer b.mtx.Unlock()

	b.subscriptions = append(b.subscriptions, sub)

	return sub
}
