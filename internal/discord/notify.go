package discord

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jakeale/tonberry/internal/monitor"
	"github.com/jakeale/tonberry/internal/store"
)

// sendTimeout bounds a single channel send. Monitor invokes handlers synchronously,
// so a send stuck on a network stall would otherwise delay the next poll tick and
// graceful shutdown indefinitely.
const sendTimeout = 15 * time.Second

// Notifier delivers world status changes detected by the monitor to every guild
// channel subscribed to the affected world.
type Notifier struct {
	session *discordgo.Session
	store   store.Store
	logger  *slog.Logger
}

// NewNotifier builds a Notifier. Register its Handle method with monitor.OnChange
// to wire it up.
func NewNotifier(session *discordgo.Session, subscriptionStore store.Store, logger *slog.Logger) *Notifier {
	return &Notifier{session: session, store: subscriptionStore, logger: logger}
}

// Handle looks up subscribers for each changed world and posts a status-change embed
// to each subscribed channel. Sends happen concurrently and a failure sending to one
// channel is logged and does not prevent delivery to the others - Monitor invokes
// handlers synchronously, so blocking here would delay its next poll tick.
//
// Pruning subscriptions whose channel has been deleted or is no longer accessible
// (Discord's Unknown Channel / Missing Access errors) is a reasonable follow-up but is
// intentionally not implemented yet - it would silently mutate guild data based on a
// transient-looking error, which deserves its own care.
func (notifier *Notifier) Handle(ctx context.Context, changes []monitor.StatusChange) {
	changesByWorld := make(map[string][]monitor.StatusChange, len(changes))
	for _, change := range changes {
		changesByWorld[change.World] = append(changesByWorld[change.World], change)
	}

	var wg sync.WaitGroup

	for world, worldChanges := range changesByWorld {
		subscribers, err := notifier.store.ListSubscribersForWorld(ctx, world)
		if err != nil {
			notifier.logger.Error("list subscribers for world failed", "world", world, "error", err)
			continue
		}

		for _, change := range worldChanges {
			embed := statusChangeEmbed(change)

			for _, subscriber := range subscribers {
				wg.Add(1)
				go func(world string, channelID string) {
					defer wg.Done()

					sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
					defer cancel()

					_, err := notifier.session.ChannelMessageSendEmbed(channelID, embed, discordgo.WithContext(sendCtx))
					if err != nil {
						notifier.logger.Error(
							"send status change notification failed",
							"world", world,
							"channel_id", channelID,
							"error", err,
						)
					}
				}(world, subscriber.ChannelID)
			}
		}
	}

	wg.Wait()
}
