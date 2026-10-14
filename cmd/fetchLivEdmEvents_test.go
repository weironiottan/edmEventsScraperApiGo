package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestScrapeLivForEdmEvents_Positive tests successful scraping scenarios
func TestScrapeLivForEdmEvents_Positive(t *testing.T) {
	// Get future dates for testing
	futureDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	laterFutureDate := time.Now().AddDate(0, 0, 31).Format("2006-01-02")

	tests := []struct {
		name                string
		mockHTML            string
		expectedEventCount  int
		expectedFirstArtist string
		expectedFirstClub   string
		validateAllEvents   func(t *testing.T, events []EdmEvent, serverURL string)
	}{
		{
			name:                "Single valid event",
			mockHTML:            livEventsPage(livEventCard("EVE-AAA111", "Tiësto", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-AAA111/tiesto/", "https://example.com/tiesto.jpg")),
			expectedEventCount:  1,
			expectedFirstArtist: "tiësto",
			expectedFirstClub:   "liv las vegas",
			validateAllEvents: func(t *testing.T, events []EdmEvent, serverURL string) {
				expectedDate := futureDate + "T00:00:00Z"
				if events[0].EventDate != expectedDate {
					t.Errorf("Expected date '%s', got '%s'", expectedDate, events[0].EventDate)
				}
				expectedTicketUrl := serverURL + "/las-vegas/event/EVE-AAA111/tiesto/"
				if events[0].TicketUrl != expectedTicketUrl {
					t.Errorf("Expected ticket URL '%s', got '%s'", expectedTicketUrl, events[0].TicketUrl)
				}
				if events[0].ArtistImageUrl != "https://example.com/tiesto.jpg" {
					t.Errorf("Expected image URL 'https://example.com/tiesto.jpg', got '%s'", events[0].ArtistImageUrl)
				}
			},
		},
		{
			name: "Multiple events keep page order",
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "Beltran", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-AAA111/beltran/", ""),
				livEventCard("EVE-BBB222", "Cloonee", "LIV Las Vegas", "livlasvegas", laterFutureDate, "/las-vegas/event/EVE-BBB222/cloonee/", ""),
			),
			expectedEventCount:  2,
			expectedFirstArtist: "beltran",
			expectedFirstClub:   "liv las vegas",
			validateAllEvents: func(t *testing.T, events []EdmEvent, serverURL string) {
				if events[1].ArtistName != "cloonee" {
					t.Errorf("Expected second artist 'cloonee', got '%s'", events[1].ArtistName)
				}
				if events[0].Id == events[1].Id {
					t.Error("Expected unique IDs for each event")
				}
			},
		},
		{
			name: "HTML entities in names are decoded",
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "Tiësto &#8211; New Year&#8217;s Eve", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-AAA111/tiesto-nye/", ""),
			),
			expectedEventCount:  1,
			expectedFirstArtist: "tiësto – new year’s eve",
			expectedFirstClub:   "liv las vegas",
		},
		{
			name: "LIV Beach events are skipped",
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "Riordan", "LIV Beach", "livlasvegasbeach", futureDate, "/las-vegas/event/EVE-AAA111/riordan/", ""),
				livEventCard("EVE-BBB222", "Beltran", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-BBB222/beltran/", ""),
			),
			expectedEventCount:  1,
			expectedFirstArtist: "beltran",
			expectedFirstClub:   "liv las vegas",
		},
		{
			name: "Duplicate cards for the same event are only added once",
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "Beltran", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-AAA111/beltran/", ""),
				livEventCard("EVE-AAA111", "Beltran", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-AAA111/beltran/", ""),
			),
			expectedEventCount:  1,
			expectedFirstArtist: "beltran",
			expectedFirstClub:   "liv las vegas",
		},
		{
			name: "Absolute ticket URLs are kept as-is",
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "Beltran", "LIV Las Vegas", "livlasvegas", futureDate, "https://www.livnightclub.com/las-vegas/event/EVE-AAA111/beltran/", ""),
			),
			expectedEventCount:  1,
			expectedFirstArtist: "beltran",
			expectedFirstClub:   "liv las vegas",
			validateAllEvents: func(t *testing.T, events []EdmEvent, serverURL string) {
				if events[0].TicketUrl != "https://www.livnightclub.com/las-vegas/event/EVE-AAA111/beltran/" {
					t.Errorf("Expected absolute ticket URL to be unchanged, got '%s'", events[0].TicketUrl)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newMockLivServer(http.StatusOK, tt.mockHTML)
			defer server.Close()

			events := scrapeLivForEdmEvents(server.URL + "/las-vegas/events/")

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
				if events[0].EventDate == "" {
					t.Error("Expected event date, got empty string")
				}
			}

			// Run custom validation if provided
			if tt.validateAllEvents != nil {
				tt.validateAllEvents(t, events, server.URL)
			}
		})
	}
}

// TestScrapeLivForEdmEvents_Negative tests error scenarios and edge cases
func TestScrapeLivForEdmEvents_Negative(t *testing.T) {
	pastDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	futureDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	tests := []struct {
		name               string
		statusCode         int
		mockHTML           string
		expectedEventCount int
		description        string
	}{
		{
			name:               "Server returns 500",
			statusCode:         http.StatusInternalServerError,
			mockHTML:           ``,
			expectedEventCount: 0,
			description:        "Should return no events when the server errors",
		},
		{
			name:               "Cloudflare block page",
			statusCode:         http.StatusForbidden,
			mockHTML:           `<html><head><title>Attention Required! | Cloudflare</title></head><body></body></html>`,
			expectedEventCount: 0,
			description:        "Should return no events when blocked",
		},
		{
			name:               "Page with no event cards",
			statusCode:         http.StatusOK,
			mockHTML:           livEventsPage(),
			expectedEventCount: 0,
			description:        "Should return no events when the layout changes",
		},
		{
			name:               "Empty body",
			statusCode:         http.StatusOK,
			mockHTML:           ``,
			expectedEventCount: 0,
			description:        "Should handle an empty response",
		},
		{
			name:       "Past events are filtered out",
			statusCode: http.StatusOK,
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "Old Artist", "LIV Las Vegas", "livlasvegas", pastDate, "/las-vegas/event/EVE-AAA111/old/", ""),
				livEventCard("EVE-BBB222", "New Artist", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-BBB222/new/", ""),
			),
			expectedEventCount: 1,
			description:        "Should drop events that already happened",
		},
		{
			name:       "Invalid date is skipped",
			statusCode: http.StatusOK,
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "Bad Date", "LIV Las Vegas", "livlasvegas", "not-a-date", "/las-vegas/event/EVE-AAA111/bad/", ""),
				livEventCard("EVE-BBB222", "Good Date", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-BBB222/good/", ""),
			),
			expectedEventCount: 1,
			description:        "Should skip events with unparseable dates",
		},
		{
			name:       "Missing date is skipped",
			statusCode: http.StatusOK,
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "No Date", "LIV Las Vegas", "livlasvegas", "", "/las-vegas/event/EVE-AAA111/nodate/", ""),
			),
			expectedEventCount: 0,
			description:        "Should skip events without a date",
		},
		{
			name:       "Placeholder events are filtered out",
			statusCode: http.StatusOK,
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "Special Guest", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-AAA111/special-guest/", ""),
				livEventCard("EVE-BBB222", "Closed For Private Event", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-BBB222/closed-for-private-event/", ""),
				livEventCard("EVE-CCC333", "Knock2", "LIV Las Vegas", "livlasvegas", futureDate, "/las-vegas/event/EVE-CCC333/knock2/", ""),
			),
			expectedEventCount: 1,
			description:        "Should drop 'special guest' and 'closed for private event' listings",
		},
		{
			name:       "Only LIV Beach events",
			statusCode: http.StatusOK,
			mockHTML: livEventsPage(
				livEventCard("EVE-AAA111", "Riordan", "LIV Beach", "livlasvegasbeach", futureDate, "/las-vegas/event/EVE-AAA111/riordan/", ""),
			),
			expectedEventCount: 0,
			description:        "Should return no events when only the dayclub has events",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			server := newMockLivServer(tt.statusCode, tt.mockHTML)
			defer server.Close()

			events := scrapeLivForEdmEvents(server.URL + "/las-vegas/events/")

			if len(events) != tt.expectedEventCount {
				t.Errorf("%s: Expected %d events, got %d", tt.description, tt.expectedEventCount, len(events))
			}
		})
	}
}

// Helpers

func newMockLivServer(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
}

func livEventsPage(cards ...string) string {
	return `<html><body><div class="lvw-events-view lvw-events-view-agenda"><div class="lvw-agenda-default">` +
		strings.Join(cards, "") +
		`</div></div></body></html>`
}

// Mirrors the card markup on https://www.livnightclub.com/las-vegas/events/
func livEventCard(eventId, name, venueName, venueKey, date, href, imageUrl string) string {
	return fmt.Sprintf(`<a class="lvw-event-list-item lvw-venuekey-%[4]s lvw-event-agenda-box livlv-event-card"
	href="%[6]s" data-eventcode="%[1]s"
	data-liv-event-id="%[1]s"
	data-liv-event-name="%[2]s"
	data-liv-venue-name="%[3]s"
	data-liv-venue-key="%[4]s"
	data-liv-event-date="%[5]s"
	aria-label="%[2]s">
		<img decoding="async" src="%[7]s" alt="%[2]s - Flyer" loading="lazy">
		<div class="livlv-event-related-card__meta"><div><h3>%[2]s</h3><small>%[3]s | 10:30pm</small></div></div>
	</a>`, eventId, name, venueName, venueKey, date, href, imageUrl)
}
