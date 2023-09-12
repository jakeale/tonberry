# tonberry

A Discord bot for FFXIV servers. Tracks Lodestone world status and notifies
subscribed channels on change, and looks up characters/Free Companies via
[godestone](https://github.com/xivapi/godestone).

## Features

- `/subscribe <world>`, `/unsubscribe <world>`, `/subscriptions` - notify a channel
  when a world's status or character-creation state changes.
- `/status [world]` - current status of one world, or a summary of all worlds.
- `/character <name> <world>` - character profile lookup.
- `/freecompany <name> <world>` - Free Company profile lookup.

## Architecture

- `internal/lodestone` - scrapes the Lodestone world-status page (goquery).
- `internal/monitor` - polls the scraper, diffs against the previous snapshot in
  memory, fans out changes. Never touches Postgres.
- `internal/store` - Postgres subscription persistence (guild/channel/world).
- `internal/godestone` - wraps `godestone/v2` for character/FC lookups, with a
  short-lived in-memory TTL cache.
- `internal/discord` - slash commands, interaction handling, embeds, notifier.

Postgres holds only durable subscription data; the world-status snapshot is small
and ephemeral, so it lives in memory.

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `DISCORD_TOKEN` | yes | - | Discord bot token |
| `DISCORD_APP_ID` | yes | - | Discord application ID |
| `DISCORD_GUILD_ID` | no | (global) | Guild to register commands to; omit for global |
| `DATABASE_URL` | yes | - | Postgres connection string |
| `LODESTONE_URL` | no | NA world-status page | Override for testing |
| `POLL_INTERVAL` | no | `30s` | World-status poll interval |
| `LOG_LEVEL` | no | `info` | `debug`, `info`, `warn`, or `error` |

## Local development

Requires Go 1.26+, Docker, and a Discord bot token.

```sh
cp .env.example .env       # fill in DISCORD_TOKEN, DISCORD_APP_ID, DISCORD_GUILD_ID
make compose-up            # Postgres on localhost:5432
export $(grep -v '^#' .env | xargs)
make migrate-up
make run
```

Set `DISCORD_GUILD_ID` to a test guild for instant command registration (global
registration can take up to an hour). Command handlers live in
`internal/discord/handlers_*.go`; restarting picks up changes.

Or run everything in containers: `docker compose up --build`.

## Testing

```sh
make test               # unit tests, no external deps
make compose-up
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/tonberry?sslmode=disable
make test-integration   # requires Postgres; skips itself if DATABASE_URL is unset
make lint                # golangci-lint (catches things go vet doesn't, e.g. errcheck)
```

CI runs all three against a Postgres service container on every push/PR.
