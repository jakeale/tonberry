package lodestone

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
)

func TestScrapeServerStatus(t *testing.T) {
	fixture, err := os.ReadFile("testdata/worldstatus.html")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(fixture); err != nil {
			t.Errorf("write fixture response: %v", err)
		}
	}))
	defer server.Close()

	scraper := NewScraper(nil, server.URL)

	servers, err := scraper.Scrape(context.Background())
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}

	if len(servers) == 0 {
		t.Fatal("expected at least one server, got none")
	}

	// "Preferred+" is a newer category not present when the original scraper was written;
	// the fixture (captured live) confirms it is now a valid value alongside the originals.
	categoryPattern := regexp.MustCompile(`^(Standard|Preferred|Preferred\+|Congested|New)$`)
	statusPattern := regexp.MustCompile(`^(Online|Offline)$`)
	creationPattern := regexp.MustCompile(
		`^Creation of New Characters (Available|Unavailable)$`,
	)

	for name, server := range servers {
		if name == "" {
			t.Error("found server with empty name")
		}

		if !categoryPattern.MatchString(server.Category) {
			t.Errorf("%s: unexpected category %q", name, server.Category)
		}

		if !statusPattern.MatchString(server.Status) {
			t.Errorf("%s: unexpected status %q", name, server.Status)
		}

		if !creationPattern.MatchString(server.CharacterCreationStatus) {
			t.Errorf("%s: unexpected characterCreationStatus %q", name, server.CharacterCreationStatus)
		}
	}
}

func TestScrapeServerStatus_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := NewScraper(nil, server.URL)

	_, err := scraper.Scrape(context.Background())
	if err == nil {
		t.Fatal("expected an error for a non-200 response, got nil")
	}
}
