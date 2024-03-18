# Start the bot
run:
	go run cmd/trading-bot/main.go

# Detect linting errors
lint:
	${GOPATH}/bin/golangci-lint run

# Run the tests
test:
	go test ./...

# Start sqlite browser (in Docker to avoid polluting your computer)
# It will be available on http://localhost:3000
start-sqlitebrowser:
	docker run -d \
	-v "${PWD}/data:/data" \
	-v "${PWD}/data/config:/config" \
	-p 3000:3000 \
	-e PUID=1000 \
	-e PGID=1000 \
	-e TZ=Europe/Paris \
	--name trading-bot-sqlitebrowser \
	--rm \
	ghcr.io/linuxserver/sqlitebrowser

# Stop sqlite browser
stop-sqlitebrowser:
	docker stop trading-bot-sqlitebrowser