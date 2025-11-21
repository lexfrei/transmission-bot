// Package bot provides the Telegram bot functionality for managing Transmission.
package bot

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/lexfrei/transmission-bot/internal/config"
	"github.com/lexfrei/transmission-bot/internal/transmission"
)

var magnetRegex = regexp.MustCompile(`magnet:\?xt=urn:[a-zA-Z0-9]+:[a-zA-Z0-9]+[^\s]*`)

// Bot represents the Telegram bot instance.
type Bot struct {
	api          *tgbotapi.BotAPI
	trClient     *transmission.Client
	allowedUsers map[int64]struct{}
	logger       *slog.Logger
}

// New creates a new Bot instance with the given configuration.
func New(cfg *config.Config, logger *slog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot: %w", err)
	}

	trClient, err := transmission.NewClient(cfg.Transmission)
	if err != nil {
		return nil, fmt.Errorf("creating transmission client: %w", err)
	}

	allowedUsers := make(map[int64]struct{}, len(cfg.Telegram.AllowedUsers))
	for _, userID := range cfg.Telegram.AllowedUsers {
		allowedUsers[userID] = struct{}{}
	}

	return &Bot{
		api:          api,
		trClient:     trClient,
		allowedUsers: allowedUsers,
		logger:       logger,
	}, nil
}

// Run starts the bot and blocks until the context is cancelled.
func (b *Bot) Run(ctx context.Context) error {
	registerErr := b.registerCommands()
	if registerErr != nil {
		return fmt.Errorf("registering commands: %w", registerErr)
	}

	b.logger.Info("bot started", "username", b.api.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := b.api.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("shutting down bot")
			b.api.StopReceivingUpdates()

			closeErr := b.trClient.Close()
			if closeErr != nil {
				b.logger.Error("failed to close transmission client", "error", closeErr)
			}

			return nil
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			go b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) registerCommands() error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "help", Description: "Show help message"},
		{Command: "list", Description: "List all torrents"},
		{Command: "remove", Description: "Remove torrent by ID"},
	}

	cfg := tgbotapi.NewSetMyCommands(commands...)

	_, err := b.api.Request(cfg)
	if err != nil {
		return fmt.Errorf("setting commands: %w", err)
	}

	b.logger.Debug("commands registered")

	return nil
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	msg := update.Message
	userID := msg.From.ID

	if !b.isAllowed(userID) {
		b.logger.Warn("unauthorized access attempt",
			"user_id", userID,
			"username", msg.From.UserName,
		)

		return
	}

	b.logger.Debug("received message",
		"user_id", userID,
		"text", msg.Text,
		"has_document", msg.Document != nil,
	)

	if msg.IsCommand() {
		b.handleCommand(ctx, msg)

		return
	}

	if msg.Document != nil {
		b.handleDocument(ctx, msg)

		return
	}

	if magnets := magnetRegex.FindAllString(msg.Text, -1); len(magnets) > 0 {
		b.handleMagnets(ctx, msg, magnets)

		return
	}
}

func (b *Bot) isAllowed(userID int64) bool {
	_, ok := b.allowedUsers[userID]

	return ok
}

func (b *Bot) handleDocument(ctx context.Context, msg *tgbotapi.Message) {
	doc := msg.Document
	if !strings.HasSuffix(strings.ToLower(doc.FileName), ".torrent") {
		b.reply(msg, "Please send a .torrent file")

		return
	}

	fileURL, err := b.api.GetFileDirectURL(doc.FileID)
	if err != nil {
		b.logger.Error("failed to get file URL", "error", err)
		b.reply(msg, "Failed to download file")

		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		b.logger.Error("failed to create request", "error", err)
		b.reply(msg, "Failed to download file")

		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		b.logger.Error("failed to download file", "error", err)
		b.reply(msg, "Failed to download file")

		return
	}

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		b.logger.Error("failed to read file", "error", err)
		b.reply(msg, "Failed to read file")

		return
	}

	base64Data := base64.StdEncoding.EncodeToString(data)

	torrent, err := b.trClient.AddTorrentByFile(ctx, base64Data)
	if err != nil {
		b.logger.Error("failed to add torrent", "error", err)
		b.reply(msg, fmt.Sprintf("Failed to add torrent: %v", err))

		return
	}

	b.logger.Info("torrent added",
		"id", torrent.ID,
		"name", torrent.Name,
		"user_id", msg.From.ID,
	)

	b.reply(msg, fmt.Sprintf("Torrent added:\nID: %d\nName: %s", torrent.ID, torrent.Name))
}

func (b *Bot) handleMagnets(ctx context.Context, msg *tgbotapi.Message, magnets []string) {
	results := make([]string, 0, len(magnets))

	for _, magnet := range magnets {
		torrent, err := b.trClient.AddTorrentByMagnet(ctx, magnet)
		if err != nil {
			b.logger.Error("failed to add magnet", "error", err)
			results = append(results, fmt.Sprintf("Failed: %v", err))

			continue
		}

		b.logger.Info("torrent added",
			"id", torrent.ID,
			"name", torrent.Name,
			"user_id", msg.From.ID,
		)

		results = append(results, fmt.Sprintf("ID: %d - %s", torrent.ID, torrent.Name))
	}

	b.reply(msg, fmt.Sprintf("Added %d torrent(s):\n%s", len(magnets), strings.Join(results, "\n")))
}

func (b *Bot) reply(msg *tgbotapi.Message, text string) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyToMessageID = msg.MessageID

	_, sendErr := b.api.Send(reply)
	if sendErr != nil {
		b.logger.Error("failed to send reply", "error", sendErr)
	}
}
