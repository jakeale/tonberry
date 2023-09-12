package godestone

import (
	"testing"
	"time"
)

func TestTTLCache_HitAndMiss(t *testing.T) {
	cache := newTTLCache[string, int](time.Minute, 10, nil)

	if _, found := cache.get("missing"); found {
		t.Error("expected miss for a key that was never set")
	}

	cache.set("a", 1)

	value, found := cache.get("a")
	if !found {
		t.Fatal("expected hit for a key that was set")
	}
	if value != 1 {
		t.Errorf("got value %d, want 1", value)
	}
}

func TestTTLCache_Expiry(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }

	cache := newTTLCache[string, int](time.Minute, 10, clock)
	cache.set("a", 1)

	now = now.Add(30 * time.Second)
	if _, found := cache.get("a"); !found {
		t.Error("expected hit before TTL elapses")
	}

	now = now.Add(31 * time.Second)
	if _, found := cache.get("a"); found {
		t.Error("expected miss after TTL elapses")
	}
}

func TestTTLCache_MaxSizeClearsOldEntries(t *testing.T) {
	cache := newTTLCache[string, int](time.Minute, 2, nil)

	cache.set("a", 1)
	cache.set("b", 2)
	cache.set("c", 3) // exceeds maxSize, clears the cache before inserting

	if _, found := cache.get("a"); found {
		t.Error("expected \"a\" to be cleared once maxSize was exceeded")
	}
	if _, found := cache.get("b"); found {
		t.Error("expected \"b\" to be cleared once maxSize was exceeded")
	}

	value, found := cache.get("c")
	if !found || value != 3 {
		t.Error("expected the entry that triggered the clear to still be present")
	}
}
