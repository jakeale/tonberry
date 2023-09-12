// Package godestone wraps github.com/xivapi/godestone/v2 to provide name+world lookups
// for characters and Free Companies, backed by a short-lived in-memory cache.
package godestone

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/karashiiro/bingode"
	upstream "github.com/xivapi/godestone/v2"
)

// ErrNotFound is returned when a search yields no exact name+world match.
var ErrNotFound = errors.New("no exact match found")

const characterCacheTTL = 5 * time.Minute
const freeCompanyCacheTTL = 10 * time.Minute
const cacheMaxSize = 256

// Client looks up characters and Free Companies on the Lodestone via godestone.
type Client struct {
	scraper *upstream.Scraper

	characterCache   *ttlCache[uint32, *upstream.Character]
	freeCompanyCache *ttlCache[string, *upstream.FreeCompany]
}

// NewClient builds a Client backed by the bingode data provider and the NA locale.
func NewClient() *Client {
	return &Client{
		scraper:          upstream.NewScraper(bingode.New(), upstream.EN),
		characterCache:   newTTLCache[uint32, *upstream.Character](characterCacheTTL, cacheMaxSize, nil),
		freeCompanyCache: newTTLCache[string, *upstream.FreeCompany](freeCompanyCacheTTL, cacheMaxSize, nil),
	}
}

// FindCharacter searches for a character by exact name and world, then fetches their
// full profile. Returns ErrNotFound if no exact match exists.
func (client *Client) FindCharacter(ctx context.Context, name, world string) (*upstream.Character, error) {
	id, err := client.searchCharacterID(ctx, name, world)
	if err != nil {
		return nil, err
	}

	if character, found := client.characterCache.get(id); found {
		return character, nil
	}

	character, err := client.scraper.FetchCharacter(id)
	if err != nil {
		return nil, err
	}

	client.characterCache.set(id, character)
	return character, nil
}

// FindFreeCompany searches for a Free Company by exact name and world, then fetches its
// full profile. Returns ErrNotFound if no exact match exists.
func (client *Client) FindFreeCompany(ctx context.Context, name, world string) (*upstream.FreeCompany, error) {
	id, err := client.searchFreeCompanyID(ctx, name, world)
	if err != nil {
		return nil, err
	}

	if freeCompany, found := client.freeCompanyCache.get(id); found {
		return freeCompany, nil
	}

	freeCompany, err := client.scraper.FetchFreeCompany(id)
	if err != nil {
		return nil, err
	}

	client.freeCompanyCache.set(id, freeCompany)
	return freeCompany, nil
}

func (client *Client) searchCharacterID(ctx context.Context, name, world string) (uint32, error) {
	results := client.scraper.SearchCharacters(upstream.CharacterOptions{Name: name, World: world})
	return searchID(ctx, results, name, world, func(result *upstream.CharacterSearchResult) (uint32, string, string, error) {
		return result.ID, result.Name, result.World, result.Error
	})
}

func (client *Client) searchFreeCompanyID(ctx context.Context, name, world string) (string, error) {
	results := client.scraper.SearchFreeCompanies(upstream.FreeCompanyOptions{Name: name, World: world})
	return searchID(ctx, results, name, world, func(result *upstream.FreeCompanySearchResult) (string, string, string, error) {
		return result.ID, result.Name, result.World, result.Error
	})
}

// searchID drains a godestone search-result channel for the first result whose name and
// world match exactly, returning ErrNotFound once the channel closes without a match.
// extract pulls the (id, name, world, error) tuple out of a single result, since the
// character and Free Company result types share no common interface.
func searchID[K comparable, R any](
	ctx context.Context,
	results <-chan R,
	wantName, wantWorld string,
	extract func(R) (id K, name, world string, err error),
) (K, error) {
	defer drainResults(results)

	var zero K
	for {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case result, open := <-results:
			if !open {
				return zero, ErrNotFound
			}
			id, name, world, err := extract(result)
			if err != nil {
				continue
			}
			if matchesNameAndWorld(name, world, wantName, wantWorld) {
				return id, nil
			}
		}
	}
}

func matchesNameAndWorld(resultName, resultWorld, wantName, wantWorld string) bool {
	return strings.EqualFold(resultName, wantName) && strings.EqualFold(resultWorld, wantWorld)
}

// drainResults consumes any remaining results after a match or context cancellation, so
// the search goroutine's send on results does not block forever.
func drainResults[T any](results <-chan T) {
	go func() {
		for range results {
		}
	}()
}
