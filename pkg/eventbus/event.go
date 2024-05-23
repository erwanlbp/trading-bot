package eventbus

type Event struct {
	Name    string
	Payload interface{}
}

const (
	EventCoinsPricesFetched    string = "coins_prices_fetched"
	EventSearchedJump          string = "searched_jump"
	EventFoundUnexistingSymbol string = "found_unexisting_symbol"
	SendNotification           string = "send_notification"
	SaveBalance                string = "save_balance"
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
