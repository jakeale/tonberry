package config

import (
	"strings"
	"testing"
	"time"
)

// setRequiredEnv sets the three required env vars to valid values so a test can
// override or unset just the one it cares about.
func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("DISCORD_TOKEN", "token")
	t.Setenv("DISCORD_APP_ID", "app-id")
	t.Setenv("DATABASE_URL", "postgres://localhost/tonberry")
}

func TestLoad_MissingRequiredVars(t *testing.T) {
	tests := []struct {
		name   string
		unset  string
		wantIn string
	}{
		{name: "missing token", unset: "DISCORD_TOKEN", wantIn: "DISCORD_TOKEN"},
		{name: "missing app id", unset: "DISCORD_APP_ID", wantIn: "DISCORD_APP_ID"},
		{name: "missing database url", unset: "DATABASE_URL", wantIn: "DATABASE_URL"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv(test.unset, "")

			_, err := Load()
			if err == nil {
				t.Fatal("Load() error = nil, want an error naming the missing var")
			}
			if got := err.Error(); !strings.Contains(got, test.wantIn) {
				t.Errorf("Load() error = %q, want it to mention %q", got, test.wantIn)
			}
		})
	}
}

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("LODESTONE_URL", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DISCORD_GUILD_ID", "")
	t.Setenv("POLL_INTERVAL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.LodestoneURL != defaultLodestoneURL {
		t.Errorf("LodestoneURL = %q, want default %q", cfg.LodestoneURL, defaultLodestoneURL)
	}
	if cfg.LogLevel != defaultLogLevel {
		t.Errorf("LogLevel = %q, want default %q", cfg.LogLevel, defaultLogLevel)
	}
	if cfg.PollInterval != defaultPollInterval {
		t.Errorf("PollInterval = %v, want default %v", cfg.PollInterval, defaultPollInterval)
	}
	if cfg.DiscordGuildID != "" {
		t.Errorf("DiscordGuildID = %q, want empty (global command registration)", cfg.DiscordGuildID)
	}
}

func TestLoad_OverridesDefaults(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("LODESTONE_URL", "https://example.test/worldstatus/")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DISCORD_GUILD_ID", "guild-123")
	t.Setenv("POLL_INTERVAL", "45s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.LodestoneURL != "https://example.test/worldstatus/" {
		t.Errorf("LodestoneURL = %q, want the override", cfg.LodestoneURL)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.DiscordGuildID != "guild-123" {
		t.Errorf("DiscordGuildID = %q, want %q", cfg.DiscordGuildID, "guild-123")
	}
	if cfg.PollInterval != 45*time.Second {
		t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, 45*time.Second)
	}
}

func TestLoad_InvalidPollIntervalReturnsError(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("POLL_INTERVAL", "not-a-duration")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want a parse error for an invalid POLL_INTERVAL")
	}
}
