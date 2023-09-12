// Package monitor polls the Lodestone world-status scraper on an interval, detects
// changes between consecutive scrapes, and fans out those changes to registered handlers.
package monitor

import (
	"context"
	"log/slog"
	"maps"
	"sync"
	"time"

	"github.com/jakeale/tonberry/internal/lodestone"
)

// ChangeHandler is invoked with the set of changes detected on a single refresh.
// It is always called with a non-empty slice.
type ChangeHandler func(ctx context.Context, changes []StatusChange)

// Monitor holds the current world-status snapshot in memory and notifies handlers
// when it changes.
type Monitor struct {
	scraper  *lodestone.Scraper
	interval time.Duration
	logger   *slog.Logger

	mutex sync.RWMutex
	// current is the last-known snapshot, exposed to callers via Snapshot.
	current lodestone.Servers

	handlersMutex sync.Mutex
	handlers      []ChangeHandler
}

// New builds a Monitor that scrapes with the given scraper on the given interval.
func New(scraper *lodestone.Scraper, interval time.Duration, logger *slog.Logger) *Monitor {
	return &Monitor{
		scraper:  scraper,
		interval: interval,
		logger:   logger,
	}
}

// OnChange registers a handler to be invoked whenever a refresh detects changes.
// Handlers are invoked synchronously and in registration order; a slow handler
// delays the next tick, so handlers should offload long work themselves.
func (monitor *Monitor) OnChange(handler ChangeHandler) {
	monitor.handlersMutex.Lock()
	defer monitor.handlersMutex.Unlock()

	monitor.handlers = append(monitor.handlers, handler)
}

// Run performs an initial scrape to seed the snapshot, then polls at the configured
// interval until ctx is canceled. No changes are reported for the initial scrape,
// matching the expectation that startup should not trigger a flood of notifications.
func (monitor *Monitor) Run(ctx context.Context) error {
	initial, err := monitor.scraper.Scrape(ctx)
	if err != nil {
		return err
	}

	monitor.mutex.Lock()
	monitor.current = initial
	monitor.mutex.Unlock()

	ticker := time.NewTicker(monitor.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			monitor.refresh(ctx)
		}
	}
}

// refresh scrapes the current world status, diffs it against the previous snapshot,
// and notifies handlers if anything changed. A failed scrape is logged and otherwise
// ignored - the previous snapshot is left untouched so the next tick retries cleanly
// rather than diffing against an empty map.
func (monitor *Monitor) refresh(ctx context.Context) {
	next, err := monitor.scraper.Scrape(ctx)
	if err != nil {
		monitor.logger.Error("scrape world status failed", "error", err)
		return
	}

	monitor.mutex.Lock()
	prev := monitor.current
	monitor.current = next
	monitor.mutex.Unlock()

	changes := diffServers(prev, next)
	if len(changes) == 0 {
		return
	}

	monitor.notify(ctx, changes)
}

func (monitor *Monitor) notify(ctx context.Context, changes []StatusChange) {
	monitor.handlersMutex.Lock()
	handlers := append([]ChangeHandler(nil), monitor.handlers...)
	monitor.handlersMutex.Unlock()

	for _, handler := range handlers {
		handler(ctx, changes)
	}
}

// Snapshot returns a copy of the most recent world-status scrape.
func (monitor *Monitor) Snapshot() lodestone.Servers {
	monitor.mutex.RLock()
	defer monitor.mutex.RUnlock()

	return maps.Clone(monitor.current)
}
