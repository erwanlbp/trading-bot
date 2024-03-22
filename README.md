# trading-bot

## Installation

- Copy `config/config.yaml.example` to `config/config.yaml` and fill it with your coins and config (each field should be detailed in the example file)

## Development

- Have Go installed
- Create a test account and API Key on [Binance Testnet](https://testnet.binance.vision/)
- Add the API Key/Secret to your `config.yaml`
- Activate `test_mode` in the `config.yaml`

Install Go dependancies

```bash
go mod tidy
```

Start the bot

```bash
make run
```

Other commands can be found in the [Makefile](Makefile)

## Deployment

- `git pull` on your server
- `go mod tidy`
- `make build`
- `nohup ./trading-bot > log.txt &` to start the bot in a backend process unlinked to your session
- `tail -f log.txt` to follow the logs
