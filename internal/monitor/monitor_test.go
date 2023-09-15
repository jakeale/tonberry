package monitor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jakeale/tonberry/internal/lodestone"
)

// fakeScraper returns scripted results in order, one per call to Scrape.
type fakeScraper struct {
	results []scrapeResult
	calls   int
}

type scrapeResult struct {
	servers lodestone.Servers
	err     error
}

func newFakeScraper(results ...scrapeResult) *fakeScraper {
	return &fakeScraper{results: results}
}

func (fake *fakeScraper) Scrape(_ context.Context) (lodestone.Servers, error) {
	result := fake.results[fake.calls]
	fake.calls++
	return result.servers, result.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// This test is in package monitor (not monitor_test) so it can drive refresh()
// directly instead of a real ticker - that keeps most cases below synchronous and
// deterministic, with no goroutines or timing to get right.

func TestMonitor_Run_InitialScrapeSeedsWithoutNotifying(t *testing.T) {
	fake := newFakeScraper(scrapeResult{
		servers: lodestone.Servers{"Adamantoise": {Status: "Online"}},
	})
	mon := New(fake, time.Hour, testLogger())

	var handlerCalled bool
	mon.OnChange(func(_ context.Context, _ []StatusChange) { handlerCalled = true })

	// The initial scrape happens unconditionally before Run checks ctx.Done, so
	// canceling up front still exercises seeding while keeping Run's call synchronous.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := mon.Run(ctx); err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}

	if handlerCalled {
		t.Error("handler was called on the initial scrape, want no notification")
	}

	if got := mon.Snapshot()["Adamantoise"].Status; got != "Online" {
		t.Errorf("Snapshot()[Adamantoise].Status = %q, want %q", got, "Online")
	}
}

func TestMonitor_Run_InitialScrapeErrorReturnsError(t *testing.T) {
	scrapeErr := errors.New("boom")
	fake := newFakeScraper(scrapeResult{err: scrapeErr})
	mon := New(fake, time.Hour, testLogger())

	err := mon.Run(context.Background())
	if !errors.Is(err, scrapeErr) {
		t.Errorf("Run() = %v, want %v", err, scrapeErr)
	}
}

func TestMonitor_Refresh_ChangedScrapeNotifiesOnce(t *testing.T) {
	fake := newFakeScraper(scrapeResult{
		servers: lodestone.Servers{"Adamantoise": {Status: "Offline"}},
	})
	mon := New(fake, time.Hour, testLogger())
	mon.current = lodestone.Servers{"Adamantoise": {Status: "Online"}}

	var received []StatusChange
	var calls int
	mon.OnChange(func(_ context.Context, changes []StatusChange) {
		calls++
		received = changes
	})

	mon.refresh(context.Background())

	if calls != 1 {
		t.Fatalf("handler called %d times, want exactly 1", calls)
	}

	want := StatusChange{World: "Adamantoise", Field: fieldStatus, OldValue: "Online", NewValue: "Offline"}
	if len(received) != 1 || received[0] != want {
		t.Errorf("changes = %+v, want [%+v]", received, want)
	}
}

func TestMonitor_Refresh_ScrapeErrorPreservesSnapshot(t *testing.T) {
	mon := New(newFakeScraper(scrapeResult{err: errors.New("boom")}), time.Hour, testLogger())
	mon.current = lodestone.Servers{"Adamantoise": {Status: "Online"}}

	var calls int
	mon.OnChange(func(_ context.Context, _ []StatusChange) { calls++ })

	mon.refresh(context.Background())

	if calls != 0 {
		t.Errorf("handler called %d times on a failed scrape, want 0", calls)
	}
	if got := mon.Snapshot()["Adamantoise"].Status; got != "Online" {
		t.Errorf("Snapshot()[Adamantoise].Status = %q after a failed scrape, want preserved %q", got, "Online")
	}

	// The next good scrape should diff against the preserved snapshot, not an empty one.
	mon.scraper = newFakeScraper(scrapeResult{servers: lodestone.Servers{"Adamantoise": {Status: "Offline"}}})
	mon.refresh(context.Background())

	if calls != 1 {
		t.Errorf("handler called %d times after the recovered scrape, want exactly 1", calls)
	}
}

func TestMonitor_Run_ContextCancellationStopsCleanly(t *testing.T) {
	fake := newFakeScraper(scrapeResult{servers: lodestone.Servers{}})
	mon := New(fake, time.Hour, testLogger())

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- mon.Run(ctx) }()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() = %v, want nil after context cancellation", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not return after context cancellation")
	}
}
