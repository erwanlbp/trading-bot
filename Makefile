run-trading-bot:
	go run cmd/trading-bot/main.go

lint:
	${GOPATH}/bin/golangci-lint run

test:
	go test ./...