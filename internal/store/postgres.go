package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the "pgx5" migrate driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all pending migrations embedded in this package to the
// database at dsn. It is safe to call on every startup: a database already at the
// latest version is left untouched.
func RunMigrations(dsn string) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", source, toPgxMigrateDSN(dsn))
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		if sourceErr, dbErr := migrator.Close(); sourceErr != nil || dbErr != nil {
			fmt.Fprintf(os.Stderr, "close migrator: source=%v database=%v\n", sourceErr, dbErr)
		}
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// toPgxMigrateDSN rewrites a standard postgres:// DSN to the pgx5:// scheme that
// golang-migrate's pgx/v5 driver expects.
func toPgxMigrateDSN(dsn string) string {
	return "pgx5://" + trimScheme(dsn)
}

func trimScheme(dsn string) string {
	dsn = strings.TrimPrefix(dsn, "postgres://")
	return strings.TrimPrefix(dsn, "postgresql://")
}

// PostgresStore is a pgx-backed implementation of Store.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore connects to Postgres and verifies the connection is usable.
func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

// Close releases the underlying connection pool.
func (store *PostgresStore) Close() {
	store.pool.Close()
}

func (store *PostgresStore) AddSubscription(ctx context.Context, guildID, channelID, worldName string) (bool, error) {
	tag, err := store.pool.Exec(ctx, `
		INSERT INTO subscriptions (guild_id, channel_id, world_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (guild_id, channel_id, world_name) DO NOTHING
	`, guildID, channelID, worldName)
	if err != nil {
		return false, fmt.Errorf("add subscription: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

func (store *PostgresStore) RemoveSubscription(ctx context.Context, guildID, channelID, worldName string) (bool, error) {
	tag, err := store.pool.Exec(ctx, `
		DELETE FROM subscriptions
		WHERE guild_id = $1 AND channel_id = $2 AND world_name = $3
	`, guildID, channelID, worldName)
	if err != nil {
		return false, fmt.Errorf("remove subscription: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

func (store *PostgresStore) ListSubscriptionsByGuild(ctx context.Context, guildID string) ([]Subscription, error) {
	rows, err := store.pool.Query(ctx, `
		SELECT id, guild_id, channel_id, world_name, created_at
		FROM subscriptions
		WHERE guild_id = $1
		ORDER BY world_name, channel_id
	`, guildID)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions by guild: %w", err)
	}
	defer rows.Close()

	return scanSubscriptions(rows)
}

func (store *PostgresStore) ListSubscribersForWorld(ctx context.Context, worldName string) ([]Subscription, error) {
	rows, err := store.pool.Query(ctx, `
		SELECT id, guild_id, channel_id, world_name, created_at
		FROM subscriptions
		WHERE world_name = $1
	`, worldName)
	if err != nil {
		return nil, fmt.Errorf("list subscribers for world: %w", err)
	}
	defer rows.Close()

	return scanSubscriptions(rows)
}

func scanSubscriptions(rows pgx.Rows) ([]Subscription, error) {
	subscriptions := make([]Subscription, 0)

	for rows.Next() {
		var subscription Subscription

		err := rows.Scan(
			&subscription.ID,
			&subscription.GuildID,
			&subscription.ChannelID,
			&subscription.WorldName,
			&subscription.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions: %w", err)
	}

	return subscriptions, nil
}

var _ Store = (*PostgresStore)(nil)
