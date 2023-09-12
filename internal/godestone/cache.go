package godestone

import (
	"sync"
	"time"
)

// ttlCache is a small, bounded, generic cache with per-entry expiry. It exists to avoid
// repeated Lodestone scrapes for the same character/FC lookup within a short window -
// Lodestone scraping is slow and rate-limit-sensitive, and popular targets are queried
// repeatedly by different users.
type ttlCache[K comparable, V any] struct {
	mutex   sync.Mutex
	entries map[K]ttlEntry[V]
	ttl     time.Duration
	maxSize int
	now     func() time.Time
}

type ttlEntry[V any] struct {
	value     V
	expiresAt time.Time
}

// newTTLCache builds a cache whose entries expire after ttl. maxSize bounds the number
// of entries retained; once exceeded, the whole cache is cleared rather than evicting
// individual entries - simple and sufficient for this cache's small expected size.
func newTTLCache[K comparable, V any](ttl time.Duration, maxSize int, now func() time.Time) *ttlCache[K, V] {
	if now == nil {
		now = time.Now
	}

	return &ttlCache[K, V]{
		entries: make(map[K]ttlEntry[V]),
		ttl:     ttl,
		maxSize: maxSize,
		now:     now,
	}
}

func (cache *ttlCache[K, V]) get(key K) (V, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	entry, found := cache.entries[key]
	if !found {
		var zero V
		return zero, false
	}

	if cache.now().After(entry.expiresAt) {
		delete(cache.entries, key)
		var zero V
		return zero, false
	}

	return entry.value, true
}

func (cache *ttlCache[K, V]) set(key K, value V) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if len(cache.entries) >= cache.maxSize {
		cache.entries = make(map[K]ttlEntry[V])
	}

	cache.entries[key] = ttlEntry[V]{
		value:     value,
		expiresAt: cache.now().Add(cache.ttl),
	}
}
