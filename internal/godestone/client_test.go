package godestone

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeSearchResult is a minimal stand-in for the upstream character/FC search result
// types, which share no common interface - searchID is generic over the extract
// function precisely so it can be tested without depending on either concrete type.
type fakeSearchResult struct {
	id    uint32
	name  string
	world string
	err   error
}

func extractFakeResult(result fakeSearchResult) (uint32, string, string, error) {
	return result.id, result.name, result.world, result.err
}

func TestSearchID_ReturnsFirstExactMatch(t *testing.T) {
	results := make(chan fakeSearchResult, 2)
	results <- fakeSearchResult{id: 1, name: "Wrong Name", world: "Adamantoise"}
	results <- fakeSearchResult{id: 2, name: "Y'shtola Rhul", world: "Adamantoise"}
	close(results)

	id, err := searchID(context.Background(), results, "Y'shtola Rhul", "Adamantoise", extractFakeResult)
	if err != nil {
		t.Fatalf("searchID() error = %v, want nil", err)
	}
	if id != 2 {
		t.Errorf("searchID() = %d, want 2", id)
	}
}

func TestSearchID_NoMatchReturnsErrNotFound(t *testing.T) {
	results := make(chan fakeSearchResult, 1)
	results <- fakeSearchResult{id: 1, name: "Someone Else", world: "Adamantoise"}
	close(results)

	_, err := searchID(context.Background(), results, "Y'shtola Rhul", "Adamantoise", extractFakeResult)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("searchID() error = %v, want %v", err, ErrNotFound)
	}
}

func TestSearchID_SkipsErroredResultsAndMatchesLater(t *testing.T) {
	results := make(chan fakeSearchResult, 2)
	results <- fakeSearchResult{err: errors.New("row parse failed")}
	results <- fakeSearchResult{id: 3, name: "Y'shtola Rhul", world: "Adamantoise"}
	close(results)

	id, err := searchID(context.Background(), results, "Y'shtola Rhul", "Adamantoise", extractFakeResult)
	if err != nil {
		t.Fatalf("searchID() error = %v, want nil", err)
	}
	if id != 3 {
		t.Errorf("searchID() = %d, want 3", id)
	}
}

func TestSearchID_CaseInsensitiveMatch(t *testing.T) {
	results := make(chan fakeSearchResult, 1)
	results <- fakeSearchResult{id: 5, name: "Y'SHTOLA RHUL", world: "ADAMANTOISE"}
	close(results)

	id, err := searchID(context.Background(), results, "y'shtola rhul", "adamantoise", extractFakeResult)
	if err != nil {
		t.Fatalf("searchID() error = %v, want nil", err)
	}
	if id != 5 {
		t.Errorf("searchID() = %d, want 5", id)
	}
}

func TestSearchID_ContextCancellationReturnsContextError(t *testing.T) {
	// Unbuffered and never sent to - searchID must return via ctx.Done() rather
	// than block forever waiting on the channel.
	results := make(chan fakeSearchResult)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := searchID(ctx, results, "Y'shtola Rhul", "Adamantoise", extractFakeResult)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("searchID() error = %v, want %v", err, context.Canceled)
	}
}

func TestSearchID_DrainsRemainingResultsWithoutBlockingSender(t *testing.T) {
	results := make(chan fakeSearchResult)
	senderDone := make(chan struct{})

	go func() {
		defer close(senderDone)

		results <- fakeSearchResult{id: 1, name: "Y'shtola Rhul", world: "Adamantoise"}
		// These sends happen after searchID has already returned below; drainResults
		// must keep consuming (concurrently with this test) so they don't block forever.
		results <- fakeSearchResult{id: 2, name: "Someone Else", world: "Adamantoise"}
		results <- fakeSearchResult{id: 3, name: "Another One", world: "Adamantoise"}
		close(results)
	}()

	id, err := searchID(context.Background(), results, "Y'shtola Rhul", "Adamantoise", extractFakeResult)
	if err != nil {
		t.Fatalf("searchID() error = %v, want nil", err)
	}
	if id != 1 {
		t.Errorf("searchID() = %d, want 1", id)
	}

	// If drainResults leaked (stopped consuming), the sender goroutine above would
	// block forever on its second send and senderDone would never close. Note: reading
	// from `results` here would race with drainResults' own background reader, so we
	// only observe completion via senderDone.
	select {
	case <-senderDone:
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for the sender to finish - drainResults may have stopped consuming")
	}
}
