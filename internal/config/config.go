// Package config loads and validates application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"time"
)

const defaultLodestoneURL = "https://na.finalfantasyxiv.com/lodestone/worldstatus/"
const defaultPollInterval = 30 * time.Second
const defaultLogLevel = "info"

// Config holds all runtime configuration for the bot.
type Config struct {
	DiscordToken   string
	DiscordAppID   string
	DiscordGuildID string // optional; empty means register commands globally
	DatabaseURL    string
	LodestoneURL   string
	PollInterval   time.Duration
	LogLevel       string
}

// Load reads configuration from the environment and validates required fields.
func Load() (Config, error) {
	cfg := Config{
		DiscordToken:   os.Getenv("DISCORD_TOKEN"),
		DiscordAppID:   os.Getenv("DISCORD_APP_ID"),
		DiscordGuildID: os.Getenv("DISCORD_GUILD_ID"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		LodestoneURL:   defaultLodestoneURL,
		PollInterval:   defaultPollInterval,
		LogLevel:       defaultLogLevel,
	}

	if lodestoneURL := os.Getenv("LODESTONE_URL"); lodestoneURL != "" {
		cfg.LodestoneURL = lodestoneURL
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	if rawInterval := os.Getenv("POLL_INTERVAL"); rawInterval != "" {
		interval, err := time.ParseDuration(rawInterval)
		if err != nil {
			return Config{}, fmt.Errorf("parse POLL_INTERVAL: %w", err)
		}
		cfg.PollInterval = interval
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	if cfg.DiscordToken == "" {
		return fmt.Errorf("DISCORD_TOKEN is required")
	}

	if cfg.DiscordAppID == "" {
		return fmt.Errorf("DISCORD_APP_ID is required")
	}

	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	return nil
}
