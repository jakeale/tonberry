// Command tonberry runs the FFXIV Discord bot: it monitors Lodestone world status,
// notifies subscribed channels of changes, and serves character/Free Company lookups.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jakeale/tonberry/internal/config"
	"github.com/jakeale/tonberry/internal/discord"
	"github.com/jakeale/tonberry/internal/godestone"
	"github.com/jakeale/tonberry/internal/lodestone"
	"github.com/jakeale/tonberry/internal/logging"
	"github.com/jakeale/tonberry/internal/monitor"
	"github.com/jakeale/tonberry/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "tonberry:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := logging.New(cfg.LogLevel)

	if err := store.RunMigrations(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	subscriptionStore, err := store.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer subscriptionStore.Close()

	godestoneClient := godestone.NewClient()

	scraper := lodestone.NewScraper(&http.Client{}, cfg.LodestoneURL)
	statusMonitor := monitor.New(scraper, cfg.PollInterval, logger)

	bot, err := discord.NewBot(cfg.DiscordToken, cfg.DiscordGuildID, subscriptionStore, statusMonitor, godestoneClient, logger)
	if err != nil {
		return fmt.Errorf("create discord bot: %w", err)
	}

	if err := bot.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}
	defer func() {
		if err := bot.Close(); err != nil {
			logger.Error("close discord session failed", "error", err)
		}
	}()

	notifier := discord.NewNotifier(bot.Session(), subscriptionStore, logger)
	statusMonitor.OnChange(notifier.Handle)

	logger.Info("tonberry started")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := statusMonitor.Run(ctx); err != nil {
			logger.Error("status monitor stopped", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	wg.Wait()

	return nil
}
