// Package lodestone scrapes FFXIV world-status information from the Lodestone website.
package lodestone

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ServerStatus describes a single world's current state.
type ServerStatus struct {
	Status                  string // "Online" | "Offline"
	Category                string // "Standard" | "Preferred" | "Congested" | "New"
	CharacterCreationStatus string // "Creation of New Characters Available" | "...Unavailable"
}

// Servers maps a world name (e.g. "Adamantoise") to its current status.
type Servers map[string]ServerStatus

// Scraper fetches and parses the Lodestone world-status page.
type Scraper struct {
	httpClient *http.Client
	url        string
}

// NewScraper builds a Scraper targeting the given world-status URL. If httpClient is nil,
// a client with a sane default timeout is used.
func NewScraper(httpClient *http.Client, url string) *Scraper {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &Scraper{httpClient: httpClient, url: url}
}

// Scrape fetches the world-status page and parses it into a Servers map.
func (scraper *Scraper) Scrape(ctx context.Context) (Servers, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scraper.url, nil)
	if err != nil {
		return nil, fmt.Errorf("build world status request: %w", err)
	}

	resp, err := scraper.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch world status page: %w", err)
	}
	// A close error on an already-fully-read response body carries no useful signal here.
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not fetch world status page: status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse world status page: %w", err)
	}

	servers := Servers{}

	doc.Find("li .item-list").Each(func(_ int, selection *goquery.Selection) {
		name := strings.TrimSpace(selection.Find("div .world-list__world_name p").Text())
		if name == "" {
			return
		}

		servers[name] = parseServerStatus(selection)
	})

	// A 200 response that parses to zero worlds almost always means Lodestone changed
	// its markup and our selectors no longer match, rather than every world vanishing.
	// Returning an error here routes the failure through the same path as a network
	// error, so the monitor preserves its last-known-good snapshot instead of wiping it.
	if len(servers) == 0 {
		return nil, fmt.Errorf("parsed zero worlds from world status page (layout may have changed)")
	}

	return servers, nil
}

// parseServerStatus extracts a single world's status fields from its <li> element.
func parseServerStatus(selection *goquery.Selection) ServerStatus {
	category := strings.TrimSpace(selection.Find("div .world-list__world_category p").Text())

	status := strings.TrimSpace(
		selection.Find("div .world-list__status_icon i").AttrOr("data-tooltip", ""),
	)

	characterCreationStatus := strings.TrimSpace(
		selection.Find("div .world-list__create_character i").AttrOr("data-tooltip", ""),
	)

	return ServerStatus{
		Status:                  status,
		Category:                category,
		CharacterCreationStatus: characterCreationStatus,
	}
}
