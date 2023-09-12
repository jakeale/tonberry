// Package store persists guild subscriptions to a world's status changes.
package store

import (
	"context"
	"time"
)

// Subscription is a single guild channel's subscription to a world's status changes.
type Subscription struct {
	ID        int64
	GuildID   string
	ChannelID string
	WorldName string
	CreatedAt time.Time
}

// Store persists and queries subscriptions.
type Store interface {
	// AddSubscription creates a subscription for the given guild channel and world.
	// created is false if the subscription already existed (the call is idempotent).
	AddSubscription(ctx context.Context, guildID, channelID, worldName string) (created bool, err error)

	// RemoveSubscription deletes a subscription for the given guild channel and world.
	// existed is false if there was nothing to remove.
	RemoveSubscription(ctx context.Context, guildID, channelID, worldName string) (existed bool, err error)

	// ListSubscriptionsByGuild returns all subscriptions for a guild, across all channels and worlds.
	ListSubscriptionsByGuild(ctx context.Context, guildID string) ([]Subscription, error)

	// ListSubscribersForWorld returns every subscription for a given world, used to fan out
	// a status-change notification to the right channels.
	ListSubscribersForWorld(ctx context.Context, worldName string) ([]Subscription, error)

	// Close releases any held resources (e.g. the underlying connection pool).
	Close()
}
