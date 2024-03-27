package eventbus

type Event struct {
	Name    string
	Payload interface{}
}

const (
	EventCoinsPricesFetched = "coins_prices_fetched"
	SendNotification        = "send_notification"
)

func GenerateEvent(eventName string, payload interface{}) Event {
	return Event{
		Name:    eventName,
		Payload: payload,
	}
}
