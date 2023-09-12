//go:build integration

package store

import (
	"context"
	"os"
	"testing"
)

func newTestStore(t *testing.T) *PostgresStore {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	if err := RunMigrations(dsn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	testStore, err := NewPostgresStore(context.Background(), dsn)
	if err != nil {
		t.Fatalf("new postgres store: %v", err)
	}

	t.Cleanup(func() {
		_, err := testStore.pool.Exec(context.Background(), "TRUNCATE subscriptions")
		if err != nil {
			t.Errorf("truncate subscriptions: %v", err)
		}
		testStore.Close()
	})

	return testStore
}

func TestPostgresStore_AddSubscription_Idempotent(t *testing.T) {
	testStore := newTestStore(t)
	ctx := context.Background()

	created, err := testStore.AddSubscription(ctx, "guild-1", "channel-1", "Adamantoise")
	if err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}
	if !created {
		t.Error("expected first AddSubscription to report created=true")
	}

	createdAgain, err := testStore.AddSubscription(ctx, "guild-1", "channel-1", "Adamantoise")
	if err != nil {
		t.Fatalf("AddSubscription (duplicate): %v", err)
	}
	if createdAgain {
		t.Error("expected duplicate AddSubscription to report created=false")
	}
}

func TestPostgresStore_RemoveSubscription(t *testing.T) {
	testStore := newTestStore(t)
	ctx := context.Background()

	_, err := testStore.AddSubscription(ctx, "guild-1", "channel-1", "Adamantoise")
	if err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}

	existed, err := testStore.RemoveSubscription(ctx, "guild-1", "channel-1", "Adamantoise")
	if err != nil {
		t.Fatalf("RemoveSubscription: %v", err)
	}
	if !existed {
		t.Error("expected RemoveSubscription to report existed=true")
	}

	existedAgain, err := testStore.RemoveSubscription(ctx, "guild-1", "channel-1", "Adamantoise")
	if err != nil {
		t.Fatalf("RemoveSubscription (already removed): %v", err)
	}
	if existedAgain {
		t.Error("expected second RemoveSubscription to report existed=false")
	}
}

func TestPostgresStore_ListSubscriptionsByGuild(t *testing.T) {
	testStore := newTestStore(t)
	ctx := context.Background()

	_, err := testStore.AddSubscription(ctx, "guild-1", "channel-1", "Adamantoise")
	if err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}
	_, err = testStore.AddSubscription(ctx, "guild-1", "channel-2", "Zalera")
	if err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}
	_, err = testStore.AddSubscription(ctx, "guild-2", "channel-3", "Zalera")
	if err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}

	subscriptions, err := testStore.ListSubscriptionsByGuild(ctx, "guild-1")
	if err != nil {
		t.Fatalf("ListSubscriptionsByGuild: %v", err)
	}

	if len(subscriptions) != 2 {
		t.Fatalf("expected 2 subscriptions for guild-1, got %d", len(subscriptions))
	}
}

func TestPostgresStore_ListSubscribersForWorld(t *testing.T) {
	testStore := newTestStore(t)
	ctx := context.Background()

	_, err := testStore.AddSubscription(ctx, "guild-1", "channel-1", "Zalera")
	if err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}
	_, err = testStore.AddSubscription(ctx, "guild-2", "channel-2", "Zalera")
	if err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}
	_, err = testStore.AddSubscription(ctx, "guild-1", "channel-1", "Adamantoise")
	if err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}

	subscribers, err := testStore.ListSubscribersForWorld(ctx, "Zalera")
	if err != nil {
		t.Fatalf("ListSubscribersForWorld: %v", err)
	}

	if len(subscribers) != 2 {
		t.Fatalf("expected 2 subscribers for Zalera, got %d", len(subscribers))
	}
}
