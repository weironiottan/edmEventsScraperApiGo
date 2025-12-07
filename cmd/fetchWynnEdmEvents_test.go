package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestScrapeWynnForEdmEvents_Positive tests successful scraping scenarios
func TestScrapeWynnForEdmEvents_Positive(t *testing.T) {
	// Get future date for testing (30 days from now)
	futureDate := time.Now().AddDate(0, 0, 30)
	futureDateStr := futureDate.Format("20060102") // YYYYMMDD format for URL extraction

	tests := []struct {
		name                string
		mockHTMLResponse    string
		expectedEventCount  int
		expectedFirstArtist string
		expectedFirstClub   string
		validateAllEvents   func(t *testing.T, events []EdmEvent)
	}{
		{
			name: "Single event",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Tiësto</span>
							<span class="venueurl">XS Nightclub</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr),
			expectedEventCount:  1,
			expectedFirstArtist: "tiësto",
			expectedFirstClub:   "xs nightclub",
			validateAllEvents: func(t *testing.T, events []EdmEvent) {
				if len(events) != 1 {
					return
				}
				if events[0].TicketUrl != fmt.Sprintf("https://wynnlasvegas.com/events/%s", futureDateStr) {
					t.Errorf("Expected ticket URL 'https://wynnlasvegas.com/events/%s', got '%s'", futureDateStr, events[0].TicketUrl)
				}
				if events[0].Id == "" {
					t.Error("Expected event ID to be generated")
				}
			},
		},
		{
			name: "Multiple events",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Calvin Harris</span>
							<span class="venueurl">XS Nightclub</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
						<div class="eventitem">
							<span class="uv-events-name">David Guetta</span>
							<span class="venueurl">Encore Beach Club</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr, futureDateStr),
			expectedEventCount:  2,
			expectedFirstArtist: "calvin harris",
			expectedFirstClub:   "xs nightclub",
		},
		{
			name: "Artist names are lowercased",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">STEVE AOKI</span>
							<span class="venueurl">XS Nightclub</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr),
			expectedEventCount:  1,
			expectedFirstArtist: "steve aoki",
			expectedFirstClub:   "xs nightclub",
		},
		{
			name: "Club names are lowercased",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Alesso</span>
							<span class="venueurl">ENCORE BEACH CLUB</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr),
			expectedEventCount:  1,
			expectedFirstArtist: "alesso",
			expectedFirstClub:   "encore beach club",
		},
		{
			name: "Wynn Field Club events are filtered out",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Golf Event</span>
							<span class="venueurl">Wynn Field Club</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr),
			expectedEventCount: 0,
		},
		{
			name: "Festival events are filtered out",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Music Festival</span>
							<span class="venueurl">Festival Grounds</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr),
			expectedEventCount: 0,
		},
		{
			name: "Art of the Wild events are filtered out",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Art Exhibition</span>
							<span class="venueurl">Art of the Wild Gallery</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr),
			expectedEventCount: 0,
		},
		{
			name:               "Empty HTML returns no events",
			mockHTMLResponse:   `<html><body></body></html>`,
			expectedEventCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.mockHTMLResponse))
			}))
			defer server.Close()

			// Call the function directly with test server URL
			events := scrapeWynnForEdmEvents(server.URL)

			// Verify event count
			if len(events) != tt.expectedEventCount {
				t.Errorf("Expected %d events, got %d", tt.expectedEventCount, len(events))
			}

			// Verify first event details if events exist
			if tt.expectedEventCount > 0 && len(events) > 0 {
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
				tt.validateAllEvents(t, events)
			}
		})
	}
}

// TestScrapeWynnForEdmEvents_Negative tests error scenarios and edge cases
func TestScrapeWynnForEdmEvents_Negative(t *testing.T) {
	// Get past date for testing (30 days ago)
	pastDate := time.Now().AddDate(0, 0, -30)
	pastDateStr := pastDate.Format("20060102") // YYYYMMDD format for URL extraction

	// Get future date
	futureDate := time.Now().AddDate(0, 0, 30)
	futureDateStr := futureDate.Format("20060102")

	tests := []struct {
		name               string
		mockHTMLResponse   string
		statusCode         int
		expectedEventCount int
		description        string
	}{
		{
			name:               "Server returns 500 error",
			mockHTMLResponse:   "",
			statusCode:         http.StatusInternalServerError,
			expectedEventCount: 0,
			description:        "Should handle server errors gracefully",
		},
		{
			name:               "Server returns 404 error",
			mockHTMLResponse:   "",
			statusCode:         http.StatusNotFound,
			expectedEventCount: 0,
			description:        "Should handle 404 errors gracefully",
		},
		{
			name: "Past events are filtered out",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Past Event</span>
							<span class="venueurl">XS Nightclub</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, pastDateStr),
			statusCode:         http.StatusOK,
			expectedEventCount: 0,
			description:        "Events with past dates should be filtered",
		},
		{
			name: "Mix of valid and past events",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Valid Event</span>
							<span class="venueurl">XS Nightclub</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
						<div class="eventitem">
							<span class="uv-events-name">Past Event</span>
							<span class="venueurl">Encore Beach Club</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr, pastDateStr),
			statusCode:         http.StatusOK,
			expectedEventCount: 1,
			description:        "Should only include future events",
		},
		{
			name: "Mix of valid and filtered events",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Valid Event</span>
							<span class="venueurl">XS Nightclub</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
						<div class="eventitem">
							<span class="uv-events-name">Festival Event</span>
							<span class="venueurl">Festival Grounds</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
						<div class="eventitem">
							<span class="uv-events-name">Golf Event</span>
							<span class="venueurl">Wynn Field Club</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr, futureDateStr, futureDateStr),
			statusCode:         http.StatusOK,
			expectedEventCount: 1,
			description:        "Should only include valid nightclub events",
		},
		{
			name: "Malformed HTML with no event containers",
			mockHTMLResponse: `
				<html>
					<body>
						<div class="some-other-class">Not an event</div>
					</body>
				</html>
			`,
			statusCode:         http.StatusOK,
			expectedEventCount: 0,
			description:        "Should handle HTML without proper event containers",
		},
		{
			name: "Missing artist name in HTML",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="venueurl">XS Nightclub</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr),
			statusCode:         http.StatusOK,
			expectedEventCount: 1,
			description:        "Should handle missing artist name gracefully",
		},
		{
			name: "Missing club name in HTML",
			mockHTMLResponse: fmt.Sprintf(`
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Artist Name</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/%s"></a>
						</div>
					</body>
				</html>
			`, futureDateStr),
			statusCode:         http.StatusOK,
			expectedEventCount: 1,
			description:        "Should handle missing club name gracefully",
		},
		{
			name: "Invalid URL format (cannot extract date)",
			mockHTMLResponse: `
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Test Artist</span>
							<span class="venueurl">XS Nightclub</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/invalid"></a>
						</div>
					</body>
				</html>
			`,
			statusCode:         http.StatusOK,
			expectedEventCount: 0,
			description:        "Should handle URLs with invalid date format",
		},
		{
			name: "Malformed date in URL",
			mockHTMLResponse: `
				<html>
					<body>
						<div class="eventitem">
							<span class="uv-events-name">Test Artist</span>
							<span class="venueurl">XS Nightclub</span>
							<a class="uv-btn" href="https://wynnlasvegas.com/events/99999999"></a>
						</div>
					</body>
				</html>
			`,
			statusCode:         http.StatusOK,
			expectedEventCount: 0,
			description:        "Should handle malformed dates gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.mockHTMLResponse))
			}))
			defer server.Close()

			// Call the function directly with test server URL
			events := scrapeWynnForEdmEvents(server.URL)

			// Verify event count
			if len(events) != tt.expectedEventCount {
				t.Errorf("%s: Expected %d events, got %d", tt.description, tt.expectedEventCount, len(events))
			}
		})
	}
}
