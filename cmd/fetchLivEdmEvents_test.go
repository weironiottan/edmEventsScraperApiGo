package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestScrapeLivForEdmEvents_Positive tests successful scraping scenarios
func TestScrapeLivForEdmEvents_Positive(t *testing.T) {
	// 10:30pm in Vegas is 05:30 or 06:30 the next day in UTC, which is how Speakeasy returns it
	futureDate := time.Now().AddDate(0, 0, 30)
	futureVegasDate := futureDate.Format("2006-01-02")
	futureUTC := futureDate.AddDate(0, 0, 1).Format("2006-01-02") + " 06:30:00"
	laterFutureUTC := futureDate.AddDate(0, 0, 2).Format("2006-01-02") + " 06:30:00"

	tests := []struct {
		name                string
		mockPages           []mockLivSpeakeasyPage
		expectedEventCount  int
		expectedFirstArtist string
		expectedFirstClub   string
		validateAllEvents   func(t *testing.T, events []EdmEvent)
	}{
		{
			name: "Single valid event",
			mockPages: []mockLivSpeakeasyPage{
				{events: []mockLivSpeakeasyEvent{{"EVE-AAA111", "Tiësto", "LIV Las Vegas", futureUTC, "prod/tiesto.jpg"}}},
			},
			expectedEventCount:  1,
			expectedFirstArtist: "tiësto",
			expectedFirstClub:   "liv las vegas",
			validateAllEvents: func(t *testing.T, events []EdmEvent) {
				expectedDate := futureVegasDate + "T00:00:00Z"
				if events[0].EventDate != expectedDate {
					t.Errorf("Expected Vegas date '%s', got '%s'", expectedDate, events[0].EventDate)
				}
				if events[0].TicketUrl != "https://speakeasygo.com/liv-nightclub-las-vegas/event?eid=EVE-AAA111" {
					t.Errorf("Unexpected ticket URL '%s'", events[0].TicketUrl)
				}
				if events[0].ArtistImageUrl != "https://media.speakeasygo.com/prod/tiesto.jpg" {
					t.Errorf("Unexpected image URL '%s'", events[0].ArtistImageUrl)
				}
			},
		},
		{
			name: "Missing flyer leaves image empty",
			mockPages: []mockLivSpeakeasyPage{
				{events: []mockLivSpeakeasyEvent{{"EVE-AAA111", "Beltran", "LIV Las Vegas", futureUTC, ""}}},
			},
			expectedEventCount:  1,
			expectedFirstArtist: "beltran",
			expectedFirstClub:   "liv las vegas",
			validateAllEvents: func(t *testing.T, events []EdmEvent) {
				if events[0].ArtistImageUrl != "" {
					t.Errorf("Expected empty image URL, got '%s'", events[0].ArtistImageUrl)
				}
			},
		},
		{
			name: "Multiple pages are followed",
			mockPages: []mockLivSpeakeasyPage{
				{hasMore: true, events: []mockLivSpeakeasyEvent{{"EVE-AAA111", "Beltran", "LIV Las Vegas", futureUTC, ""}}},
				{events: []mockLivSpeakeasyEvent{{"EVE-BBB222", "Cloonee", "LIV Las Vegas", laterFutureUTC, ""}}},
			},
			expectedEventCount:  2,
			expectedFirstArtist: "beltran",
			expectedFirstClub:   "liv las vegas",
			validateAllEvents: func(t *testing.T, events []EdmEvent) {
				if events[1].ArtistName != "cloonee" {
					t.Errorf("Expected second artist 'cloonee', got '%s'", events[1].ArtistName)
				}
				if events[0].Id == events[1].Id {
					t.Error("Expected unique IDs for each event")
				}
			},
		},
		{
			name: "Events repeated across pages are only added once",
			mockPages: []mockLivSpeakeasyPage{
				{hasMore: true, events: []mockLivSpeakeasyEvent{{"EVE-AAA111", "Beltran", "LIV Las Vegas", futureUTC, ""}}},
				{events: []mockLivSpeakeasyEvent{
					{"EVE-AAA111", "Beltran", "LIV Las Vegas", futureUTC, ""},
					{"EVE-BBB222", "Cloonee", "LIV Las Vegas", laterFutureUTC, ""},
				}},
			},
			expectedEventCount:  2,
			expectedFirstArtist: "beltran",
			expectedFirstClub:   "liv las vegas",
		},
		{
			name: "Other venues from the search are skipped",
			mockPages: []mockLivSpeakeasyPage{
				{events: []mockLivSpeakeasyEvent{
					{"EVE-AAA111", "Riordan", "LIV Beach", futureUTC, ""},
					{"EVE-BBB222", "DJ Somebody", "LIV Miami", futureUTC, ""},
					{"EVE-CCC333", "Beltran", "LIV Las Vegas", futureUTC, ""},
				}},
			},
			expectedEventCount:  1,
			expectedFirstArtist: "beltran",
			expectedFirstClub:   "liv las vegas",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newMockLivSpeakeasyServer(t, http.StatusOK, tt.mockPages)
			defer server.Close()

			events := scrapeLivForEdmEvents(server.URL + "/discovery?tab=events&search=LIV%20Las%20Vegas")

			// Verify event count
			if len(events) != tt.expectedEventCount {
				t.Fatalf("Expected %d events, got %d", tt.expectedEventCount, len(events))
			}

			// Verify first event details if events exist
			if tt.expectedEventCount > 0 {
				if events[0].ArtistName != tt.expectedFirstArtist {
					t.Errorf("Expected first artist '%s', got '%s'", tt.expectedFirstArtist, events[0].ArtistName)
				}
				if events[0].ClubName != tt.expectedFirstClub {
					t.Errorf("Expected first club '%s', got '%s'", tt.expectedFirstClub, events[0].ClubName)
				}
				if events[0].Id == "" {
					t.Error("Expected event ID to be generated, got empty string")
				}
			}

			// Run custom validation if provided
			if tt.validateAllEvents != nil {
				tt.validateAllEvents(t, events)
			}
		})
	}
}

// TestScrapeLivForEdmEvents_Negative tests error scenarios and edge cases
func TestScrapeLivForEdmEvents_Negative(t *testing.T) {
	pastUTC := time.Now().AddDate(0, 0, -30).Format("2006-01-02") + " 06:30:00"
	futureUTC := time.Now().AddDate(0, 0, 30).Format("2006-01-02") + " 06:30:00"

	tests := []struct {
		name               string
		statusCode         int
		rawBody            string
		mockPages          []mockLivSpeakeasyPage
		expectedEventCount int
		description        string
	}{
		{
			name:               "Server returns 500",
			statusCode:         http.StatusInternalServerError,
			rawBody:            ``,
			expectedEventCount: 0,
			description:        "Should return no events when the server errors",
		},
		{
			name:               "Cloudflare block page",
			statusCode:         http.StatusForbidden,
			rawBody:            `<html><head><title>Just a moment...</title></head><body></body></html>`,
			expectedEventCount: 0,
			description:        "Should return no events when blocked",
		},
		{
			name:               "Page without __NEXT_DATA__",
			statusCode:         http.StatusOK,
			rawBody:            `<html><body><div>Events</div></body></html>`,
			expectedEventCount: 0,
			description:        "Should return no events when the layout changes",
		},
		{
			name:               "Invalid __NEXT_DATA__ JSON",
			statusCode:         http.StatusOK,
			rawBody:            `<html><body><script id="__NEXT_DATA__" type="application/json">{not json</script></body></html>`,
			expectedEventCount: 0,
			description:        "Should handle malformed JSON",
		},
		{
			name:               "No events",
			statusCode:         http.StatusOK,
			mockPages:          []mockLivSpeakeasyPage{{events: []mockLivSpeakeasyEvent{}}},
			expectedEventCount: 0,
			description:        "Should handle an empty search",
		},
		{
			name:       "Past events are filtered out",
			statusCode: http.StatusOK,
			mockPages: []mockLivSpeakeasyPage{{events: []mockLivSpeakeasyEvent{
				{"EVE-AAA111", "Old Artist", "LIV Las Vegas", pastUTC, ""},
				{"EVE-BBB222", "New Artist", "LIV Las Vegas", futureUTC, ""},
			}}},
			expectedEventCount: 1,
			description:        "Should drop events that already happened",
		},
		{
			name:       "Invalid date is skipped",
			statusCode: http.StatusOK,
			mockPages: []mockLivSpeakeasyPage{{events: []mockLivSpeakeasyEvent{
				{"EVE-AAA111", "Bad Date", "LIV Las Vegas", "not-a-date", ""},
				{"EVE-BBB222", "Good Date", "LIV Las Vegas", futureUTC, ""},
			}}},
			expectedEventCount: 1,
			description:        "Should skip events with unparseable dates",
		},
		{
			name:       "Placeholder events are filtered out",
			statusCode: http.StatusOK,
			mockPages: []mockLivSpeakeasyPage{{events: []mockLivSpeakeasyEvent{
				{"EVE-AAA111", "Special Guest", "LIV Las Vegas", futureUTC, ""},
				{"EVE-BBB222", "Closed For Private Event", "LIV Las Vegas", futureUTC, ""},
				{"EVE-CCC333", "VIVE: Latin Night w Olea", "LIV Las Vegas", futureUTC, ""},
				{"EVE-DDD444", "Knock2", "LIV Las Vegas", futureUTC, ""},
			}}},
			expectedEventCount: 1,
			description:        "Should drop 'special guest', 'closed for private event' and 'latin night' listings",
		},
		{
			name:       "Server always says there are more pages",
			statusCode: http.StatusOK,
			mockPages: []mockLivSpeakeasyPage{
				{hasMore: true, events: []mockLivSpeakeasyEvent{{"EVE-AAA111", "Beltran", "LIV Las Vegas", futureUTC, ""}}},
			},
			expectedEventCount: 1,
			description:        "Should stop paging instead of looping forever",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var server *httptest.Server
			if tt.mockPages != nil {
				server = newMockLivSpeakeasyServer(t, tt.statusCode, tt.mockPages)
			} else {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(tt.statusCode)
					w.Write([]byte(tt.rawBody))
				}))
			}
			defer server.Close()

			events := scrapeLivForEdmEvents(server.URL + "/discovery?tab=events&search=LIV%20Las%20Vegas")

			if len(events) != tt.expectedEventCount {
				t.Errorf("%s: Expected %d events, got %d", tt.description, tt.expectedEventCount, len(events))
			}
		})
	}
}

// Helper types

type mockLivSpeakeasyEvent struct {
	uniqueId      string
	title         string
	venueName     string
	startDateTime string
	flyerKey      string
}

type mockLivSpeakeasyPage struct {
	hasMore bool
	events  []mockLivSpeakeasyEvent
}

// Serves the pages in order, repeating the last one if the scraper keeps asking
func newMockLivSpeakeasyServer(t *testing.T, statusCode int, pages []mockLivSpeakeasyPage) *httptest.Server {
	requestCount := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("startDate") == "" || r.URL.Query().Get("endDate") == "" {
			t.Errorf("Expected startDate and endDate query params, got %s", r.URL.RawQuery)
		}

		page := pages[len(pages)-1]
		if requestCount < len(pages) {
			page = pages[requestCount]
		}
		requestCount++

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(statusCode)
		w.Write([]byte(livSpeakeasyPage(t, page)))
	}))
}

// Mirrors the __NEXT_DATA__ on https://speakeasygo.com/discovery?tab=events
func livSpeakeasyPage(t *testing.T, page mockLivSpeakeasyPage) string {
	events := []map[string]any{}
	for _, e := range page.events {
		events = append(events, map[string]any{
			"uniqueId":      e.uniqueId,
			"title":         e.title,
			"startDateTime": e.startDateTime,
			"flyer":         map[string]any{"key": e.flyerKey},
			"venue":         map[string]any{"name": e.venueName, "city": "Las Vegas", "timezone": "America/Los_Angeles"},
		})
	}
	nextData, err := json.Marshal(map[string]any{
		"props": map[string]any{
			"pageProps": map[string]any{"events": events, "eventsHasMore": page.hasMore},
		},
		"page": "/discovery",
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf(`<html><body><div id="__next"></div><script id="__NEXT_DATA__" type="application/json">%s</script></body></html>`, nextData)
}
