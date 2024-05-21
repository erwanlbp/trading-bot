package eventbus

type Event struct {
	Name    string
	Payload interface{}
}

const (
	EventCoinsPricesFetched    = "coins_prices_fetched"
	EventSearchedJump          = "coins_prices_fetched"
	EventFoundUnexistingSymbol = "found_unexisting_symbol"
	SendNotification           = "send_notification"
	SaveBalance                = "save_balance"
)

func GenerateEvent(eventName string, payload interface{}) Event {
	return Event{
		Name:    eventName,
		Payload: payload,
	}
}

func FoundUnexistingSymbol(symbol string) Event {
	return GenerateEvent(EventFoundUnexistingSymbol, symbol)
}

func SearchedJump() Event {
	return GenerateEvent(EventSearchedJump, nil)
}
