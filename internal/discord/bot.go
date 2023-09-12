// Package discord wires the Discord bot: slash-command registration, interaction
// handling, and status-change notifications.
package discord

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/jakeale/tonberry/internal/godestone"
	"github.com/jakeale/tonberry/internal/monitor"
	"github.com/jakeale/tonberry/internal/store"
)

// Bot owns the Discord session and dispatches interactions to command handlers.
type Bot struct {
	session *discordgo.Session
	guildID string

	store           store.Store
	monitor         *monitor.Monitor
	godestoneClient *godestone.Client
	logger          *slog.Logger
}

// NewBot creates a Discord session for the given token, but does not open it or
// register commands - call Open for that.
func NewBot(
	token string,
	guildID string,
	subscriptionStore store.Store,
	statusMonitor *monitor.Monitor,
	godestoneClient *godestone.Client,
	logger *slog.Logger,
) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		session:         session,
		guildID:         guildID,
		store:           subscriptionStore,
		monitor:         statusMonitor,
		godestoneClient: godestoneClient,
		logger:          logger,
	}

	session.AddHandler(bot.handleInteraction)

	return bot, nil
}

// Open connects to Discord and registers slash commands.
func (bot *Bot) Open() error {
	if err := bot.session.Open(); err != nil {
		return err
	}

	appID := bot.session.State.User.ID

	_, err := bot.session.ApplicationCommandBulkOverwrite(appID, bot.guildID, commandDefinitions)
	if err != nil {
		return err
	}

	return nil
}

// Close closes the Discord session.
func (bot *Bot) Close() error {
	return bot.session.Close()
}

// Session exposes the underlying discordgo session, e.g. for the Notifier to send
// channel messages.
func (bot *Bot) Session() *discordgo.Session {
	return bot.session
}
