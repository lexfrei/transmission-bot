package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Size constants for formatting.
const (
	bytesInKilobyte = 1024
	bytesInMegabyte = bytesInKilobyte * 1024
	bytesInGigabyte = bytesInMegabyte * 1024
	bytesInTerabyte = bytesInGigabyte * 1024
	percentMultiply = 100
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
/remove <id> delete - Remove torrent and delete data

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

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Torrents (%d):\n\n", len(torrents)))

	for _, torrent := range torrents {
		builder.WriteString(fmt.Sprintf(
			"[%d] %s\n%s | %.1f%% | %s\n\n",
			torrent.ID,
			torrent.Name,
			torrent.Status,
			torrent.PercentDone*percentMultiply,
			formatSize(torrent.TotalSize),
		))
	}

	b.reply(msg, builder.String())
}

func (b *Bot) handleRemove(ctx context.Context, msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())

	if len(args) == 0 {
		b.reply(msg, "Usage: /remove <id> [delete]\n\nAdd 'delete' to also remove downloaded data.")

		return
	}

	torrentID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		b.reply(msg, "Invalid torrent ID. Please provide a numeric ID.")

		return
	}

	deleteData := len(args) > 1 && args[1] == "delete"

	removeErr := b.trClient.RemoveTorrent(ctx, torrentID, deleteData)
	if removeErr != nil {
		b.logger.Error("failed to remove torrent", "error", removeErr, "id", torrentID)
		b.reply(msg, fmt.Sprintf("Failed to remove torrent: %v", removeErr))

		return
	}

	b.logger.Info("torrent removed",
		"id", torrentID,
		"delete_data", deleteData,
		"user_id", msg.From.ID,
	)

	if deleteData {
		b.reply(msg, fmt.Sprintf("Torrent %d removed with data", torrentID))
	} else {
		b.reply(msg, fmt.Sprintf("Torrent %d removed", torrentID))
	}
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= bytesInTerabyte:
		return fmt.Sprintf("%.2f TB", float64(bytes)/bytesInTerabyte)
	case bytes >= bytesInGigabyte:
		return fmt.Sprintf("%.2f GB", float64(bytes)/bytesInGigabyte)
	case bytes >= bytesInMegabyte:
		return fmt.Sprintf("%.2f MB", float64(bytes)/bytesInMegabyte)
	case bytes >= bytesInKilobyte:
		return fmt.Sprintf("%.2f KB", float64(bytes)/bytesInKilobyte)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
