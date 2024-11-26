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
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Run the tests
test:
	go test ./...
