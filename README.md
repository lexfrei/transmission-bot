# transmission-bot

[![Go](https://img.shields.io/badge/Go-1.24-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-BSD--3--Clause-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/lexfrei/transmission-bot)](https://github.com/lexfrei/transmission-bot/releases)

Telegram bot for managing Transmission torrent client.

## Features

- Add torrents via `.torrent` files
- Add torrents via magnet links
- List active torrents with progress
- Remove torrents (with optional data deletion)
- Whitelist-based access control by Telegram user ID
- Structured logging with slog
- Configuration via YAML, environment variables, or CLI flags

## Requirements

- Go 1.24+
- Transmission 4.0+ with RPC enabled
- Telegram Bot Token (from [@BotFather](https://t.me/BotFather))

## Installation

### Container

```bash
docker pull ghcr.io/lexfrei/transmission-bot:latest
```

### From source

```bash
go install github.com/lexfrei/transmission-bot/cmd/transmission-bot@latest
```

## Configuration

### Environment variables

| Variable | Description | Default |
| -------- | ----------- | ------- |
| `TB_TELEGRAM_TOKEN` | Telegram bot token | *required* |
| `TB_TELEGRAM_ALLOWED_USERS` | Comma-separated list of allowed Telegram user IDs | *required* |
| `TB_TRANSMISSION_URL` | Transmission RPC URL | `http://localhost:9091/transmission/rpc` |
| `TB_TRANSMISSION_USERNAME` | Transmission username | *empty* |
| `TB_TRANSMISSION_PASSWORD` | Transmission password | *empty* |
| `TB_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |

### Config file

```yaml
telegram:
  token: "YOUR_BOT_TOKEN"
  allowed_users:
    - 123456789

transmission:
  url: "http://localhost:9091/transmission/rpc"
  username: ""
  password: ""

log:
  level: "info"
```

### CLI flags

```bash
transmission-bot --help
```

## Usage

### Running with Docker

```bash
docker run -d \
  --name transmission-bot \
  -e TB_TELEGRAM_TOKEN=your_token \
  -e TB_TELEGRAM_ALLOWED_USERS=123456789 \
  -e TB_TRANSMISSION_URL=http://transmission:9091/transmission/rpc \
  ghcr.io/lexfrei/transmission-bot:latest
```

### Running locally

```bash
export TB_TELEGRAM_TOKEN=your_token
export TB_TELEGRAM_ALLOWED_USERS=123456789
./transmission-bot
```

## Bot Commands

| Command | Description |
| ------- | ----------- |
| `/start` | Start the bot |
| `/help` | Show help message |
| `/list` | List all torrents |
| `/remove <id>` | Remove torrent by ID |
| `/remove <id> data` | Remove torrent and delete data |

You can also send:

- `.torrent` files to add new torrents
- Magnet links to add new torrents

## Development

### Build

```bash
go build -o transmission-bot ./cmd/transmission-bot
```

### Test

```bash
go test -race ./...
```

### Lint

```bash
golangci-lint run
```

## License

[BSD-3-Clause](LICENSE)

## Author

Aleksei Sviridkin <f@lex.la>
