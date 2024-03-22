cmd=trading-bot
build:
	go build -o ${cmd} cmd/${cmd}/main.go

# Start the bot
run:
	go run cmd/${cmd}/main.go

# Detect linting errors
lint:
	${GOPATH}/bin/golangci-lint run

# Run the tests
test:
	go test ./...
