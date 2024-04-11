cmd=trading-bot
build:
	go build -o ${cmd} cmd/${cmd}/main.go

build-all:
	go build -o trading-bot cmd/trading-bot/main.go
	go build -o balances cmd/balances/main.go

# Start the bot
run:
	go run cmd/${cmd}/main.go

# Detect linting errors
lint:
	${GOPATH}/bin/golangci-lint run

# Run the tests
test:
	go test ./...
