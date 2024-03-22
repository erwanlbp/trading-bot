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

### To start the bot on production

- `git pull` on your server
- `go mod tidy`
- `make build`
- (Check the `config.yaml`, test mode, thresholds, etc)
- `nohup ./trading-bot > log.txt &` to start the bot in a backend process unlinked to your session
- `tail -f log.txt` to follow the logs

### To stop the bot

- `jobs` to list the commands ending with `&` that are running
- `fg` to make it a "frontend" process
- `ctrl-c` to cancel the process (it will wait 2s to finish cleanly)
- `ps -aux | grep "trading-bot"` to check that there's no trading-bot process running anymore
