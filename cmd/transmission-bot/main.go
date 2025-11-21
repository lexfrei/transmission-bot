// Package main provides the entry point for the transmission-bot application.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lexfrei/transmission-bot/internal/bot"
	"github.com/lexfrei/transmission-bot/internal/config"
)

// Build-time variables set via ldflags.
//
//nolint:gochecknoglobals // Required for ldflags injection at build time
var (
	Version  = "development"
	Revision = "unknown"
)

// CLI flags and command - global variables required by cobra pattern.
//
//nolint:gochecknoglobals // Cobra requires global variables for command registration
var (
	cfgFile  string
	logLevel string
)

//nolint:gochecknoglobals // Cobra requires global command variable
var rootCmd = &cobra.Command{
	Use:     "transmission-bot",
	Short:   "Telegram bot for Transmission torrent client",
	Version: Version + " (" + Revision + ")",
	Long: `A Telegram bot that allows you to manage your Transmission
downloads. Send torrent files or magnet links to add new downloads,
list active torrents, and remove completed ones.`,
	RunE: run,
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//nolint:gochecknoinits // Cobra requires init for flag registration
func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")

	rootCmd.PersistentFlags().String("telegram-token", "", "Telegram bot token")
	rootCmd.PersistentFlags().Int64Slice("telegram-allowed-users", nil, "Allowed Telegram user IDs")
	rootCmd.PersistentFlags().String("transmission-url", "", "Transmission RPC URL")
	rootCmd.PersistentFlags().String("transmission-username", "", "Transmission username")
	rootCmd.PersistentFlags().String("transmission-password", "", "Transmission password")

	_ = viper.BindPFlag("telegram.token", rootCmd.PersistentFlags().Lookup("telegram-token"))
	_ = viper.BindPFlag("telegram.allowed_users", rootCmd.PersistentFlags().Lookup("telegram-allowed-users"))
	_ = viper.BindPFlag("transmission.url", rootCmd.PersistentFlags().Lookup("transmission-url"))
	_ = viper.BindPFlag("transmission.username", rootCmd.PersistentFlags().Lookup("transmission-username"))
	_ = viper.BindPFlag("transmission.password", rootCmd.PersistentFlags().Lookup("transmission-password"))
	_ = viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
}

func run(_ *cobra.Command, _ []string) error {
	logger := setupLogger(logLevel)

	cfg, err := config.Load(cfgFile)
	if err != nil {
		logger.Error("failed to load config", "error", err)

		return fmt.Errorf("loading config: %w", err)
	}

	logger.Info("configuration loaded",
		"transmission_url", cfg.Transmission.URL,
		"allowed_users", cfg.Telegram.AllowedUsers,
	)

	telegramBot, err := bot.New(cfg, logger)
	if err != nil {
		logger.Error("failed to create bot", "error", err)

		return fmt.Errorf("creating bot: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Info("received signal, shutting down", "signal", sig)
		cancel()
	}()

	runErr := telegramBot.Run(ctx)
	if runErr != nil {
		return fmt.Errorf("running bot: %w", runErr)
	}

	return nil
}

func setupLogger(level string) *slog.Logger {
	var slogLevel slog.Level

	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})

	return slog.New(handler)
}
