# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build
go build -o transmission-bot ./cmd/transmission-bot

# Run tests
go test -race ./...

# Lint (all linters enabled by default, see .golangci.yaml for disabled ones)
golangci-lint run

# Markdown lint
markdownlint "**/*.md"

# Run locally
export TB_TELEGRAM_TOKEN=your_token
export TB_TELEGRAM_ALLOWED_USERS=123456789
./transmission-bot
```

## Architecture

Telegram bot for managing Transmission torrent client via RPC.

```text
cmd/transmission-bot/    Entry point, CLI setup (cobra), logger configuration
internal/
  bot/                   Telegram bot: update handling, command routing, message replies
  config/                Viper-based config loading (YAML, env vars, CLI flags)
  transmission/          Wrapper around github.com/lexfrei/go-transmission client
```

### Key Dependencies

- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API
- `github.com/lexfrei/go-transmission` - Transmission RPC client
- `github.com/spf13/cobra` + `github.com/spf13/viper` - CLI and config

### Logging

Uses `log/slog` with JSON handler. Third-party library logs (telegram-bot-api) are routed through slog via adapter implementing `tgbotapi.BotLogger` interface.

### Configuration Priority

1. CLI flags
2. Environment variables (prefix `TB_`)
3. Config file (config.yaml)
4. Defaults
