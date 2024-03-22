build:
	go build -o trading-bot cmd/trading-bot/main.go

# Start the bot
run:
	go run cmd/trading-bot/main.go

# When you want to test little pieces of code, without starting the bot
run-test:
	go run cmd/test/main.go

# Detect linting errors
lint:
	${GOPATH}/bin/golangci-lint run

# Run the tests
test:
	go test ./...
