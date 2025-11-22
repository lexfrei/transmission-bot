package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	percentMultiply  = 100
	maxMessageLength = 4096
)

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.handleStart(msg)
	case "help":
		b.handleHelp(msg)
	case "list":
		b.handleList(ctx, msg)
	case "remove":
		b.handleRemove(ctx, msg)
	default:
		b.reply(msg, "Unknown command. Use /help to see available commands.")
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	text := fmt.Sprintf(
		"Welcome, %s!\n\n"+
			"I can help you manage your Transmission downloads.\n\n"+
			"Send me a .torrent file or a magnet link to add a new torrent.\n\n"+
			"Use /help to see all available commands.",
		msg.From.FirstName,
	)

	b.reply(msg, text)
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	text := `Available commands:

/start - Start the bot
/help - Show this help message
/list - List all torrents
/remove <id> - Remove torrent by ID
/remove <id> data - Remove torrent and delete data

You can also:
• Send a .torrent file
• Send a magnet link`

	b.reply(msg, text)
}

func (b *Bot) handleList(ctx context.Context, msg *tgbotapi.Message) {
	torrents, err := b.trClient.ListTorrents(ctx)
	if err != nil {
		b.logger.Error("failed to list torrents", "error", err)
		b.reply(msg, fmt.Sprintf("Failed to list torrents: %v", err))

		return
	}

	if len(torrents) == 0 {
		b.reply(msg, "No torrents found")

		return
	}

	header := fmt.Sprintf("Torrents (%d):\n", len(torrents))

	var messages []string

	var current strings.Builder

	current.WriteString(header)

	for _, torrent := range torrents {
		line := fmt.Sprintf(
			"[%d] %s - %.0f%%\n",
			torrent.ID,
			torrent.Name,
			torrent.PercentDone*percentMultiply,
		)

		if current.Len()+len(line) > maxMessageLength {
			messages = append(messages, current.String())
			current.Reset()
		}

		current.WriteString(line)
	}

	if current.Len() > 0 {
		messages = append(messages, current.String())
	}

	for _, message := range messages {
		b.reply(msg, message)
	}
}

func (b *Bot) handleRemove(ctx context.Context, msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())

	if len(args) == 0 {
		b.reply(msg, "Usage: /remove <id> [data]\n\nAdd 'data' to also remove downloaded data.")

		return
	}

	torrentID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		b.reply(msg, "Invalid torrent ID. Please provide a numeric ID.")

		return
	}

	torrent, getErr := b.trClient.GetTorrent(ctx, torrentID)
	if getErr != nil {
		b.logger.Error("failed to get torrent", "error", getErr, "id", torrentID)
		b.reply(msg, fmt.Sprintf("Failed to find torrent: %v", getErr))

		return
	}

	deleteData := len(args) > 1 && args[1] == "data"

	removeErr := b.trClient.RemoveTorrent(ctx, torrentID, deleteData)
	if removeErr != nil {
		b.logger.Error("failed to remove torrent", "error", removeErr, "id", torrentID)
		b.reply(msg, fmt.Sprintf("Failed to remove torrent: %v", removeErr))

		return
	}

	b.logger.Info("torrent removed",
		"id", torrentID,
		"name", torrent.Name,
		"delete_data", deleteData,
		"user_id", msg.From.ID,
	)

	if deleteData {
		b.reply(msg, "Removed with data: "+torrent.Name)
	} else {
		b.reply(msg, "Removed: "+torrent.Name)
	}
}
