package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestScrapeZoukEdmEvents_Positive tests successful scraping scenarios
func TestScrapeZoukEdmEvents_Positive(t *testing.T) {
	// Get future date for testing (30 days from now)
	futureDate := time.Now().AddDate(0, 0, 30)
	futureDateStr := futureDate.Format("20060102") // YYYYMMDD format for URL extraction

	tests := []struct {
		name                string
		mockHTMLResponses   []string
		expectedEventCount  int
		expectedFirstArtist string
		expectedFirstClub   string
		validateAllEvents   func(t *testing.T, events []EdmEvent)
	}{
		{
			name: "Single month with one event",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Tiësto</span>
								<a class="venueurl">AYU Dayclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				`<html><body></body></html>`, // Empty response stops pagination
			},
			expectedEventCount:  1,
			expectedFirstArtist: "tiësto",
			expectedFirstClub:   "ayu dayclub",
			validateAllEvents: func(t *testing.T, events []EdmEvent) {
				if len(events) != 1 {
					return
				}
				if events[0].TicketUrl != fmt.Sprintf("https://zoukgrouplv.com/events/%s", futureDateStr) {
					t.Errorf("Expected ticket URL 'https://zoukgrouplv.com/events/%s', got '%s'", futureDateStr, events[0].TicketUrl)
				}
				if events[0].Id == "" {
					t.Error("Expected event ID to be generated")
				}
			},
		},
		{
			name: "Single month with multiple events",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Calvin Harris</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
							<div class="eventitem">
								<span class="uv-event-name">David Guetta</span>
								<a class="venueurl">AYU Dayclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr, futureDateStr),
				`<html><body></body></html>`,
			},
			expectedEventCount:  2,
			expectedFirstArtist: "calvin harris",
			expectedFirstClub:   "zouk nightclub",
		},
		{
			name: "Multiple months with pagination",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Zedd</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Marshmello</span>
								<a class="venueurl">AYU Dayclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				`<html><body></body></html>`,
			},
			expectedEventCount:  2,
			expectedFirstArtist: "zedd",
			expectedFirstClub:   "zouk nightclub",
		},
		{
			name: "Artist names are lowercased",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">STEVE AOKI</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				`<html><body></body></html>`,
			},
			expectedEventCount:  1,
			expectedFirstArtist: "steve aoki",
			expectedFirstClub:   "zouk nightclub",
		},
		{
			name: "Club names are lowercased",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Alesso</span>
								<a class="venueurl">ZOUK NIGHTCLUB</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				`<html><body></body></html>`,
			},
			expectedEventCount:  1,
			expectedFirstArtist: "alesso",
			expectedFirstClub:   "zouk nightclub",
		},
		{
			name: "Empty HTML returns no events",
			mockHTMLResponses: []string{
				`<html><body></body></html>`,
			},
			expectedEventCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			responseIndex := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if responseIndex < len(tt.mockHTMLResponses) {
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockHTMLResponses[responseIndex]))
					responseIndex++
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			// Call the function directly with test server URL
			events := scrapeZoukEdmEvents(server.URL + "?date=")

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

// TestScrapeZoukEdmEvents_Negative tests error scenarios and edge cases
func TestScrapeZoukEdmEvents_Negative(t *testing.T) {
	// Get past date for testing (30 days ago)
	pastDate := time.Now().AddDate(0, 0, -30)
	pastDateStr := pastDate.Format("20060102") // YYYYMMDD format for URL extraction

	// Get future date
	futureDate := time.Now().AddDate(0, 0, 30)
	futureDateStr := futureDate.Format("20060102")

	tests := []struct {
		name               string
		mockHTMLResponses  []string
		statusCodes        []int
		expectedEventCount int
		description        string
	}{
		{
			name:               "Server returns 500 error",
			mockHTMLResponses:  []string{""},
			statusCodes:        []int{http.StatusInternalServerError},
			expectedEventCount: 0,
			description:        "Should handle server errors gracefully",
		},
		{
			name:               "Server returns 404 error",
			mockHTMLResponses:  []string{""},
			statusCodes:        []int{http.StatusNotFound},
			expectedEventCount: 0,
			description:        "Should handle 404 errors gracefully",
		},
		{
			name: "Past events are filtered out",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Past Event</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, pastDateStr),
				`<html><body></body></html>`,
			},
			statusCodes:        []int{http.StatusOK, http.StatusOK},
			expectedEventCount: 0,
			description:        "Events with past dates should be filtered",
		},
		{
			name: "Mix of valid and past events",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Valid Event</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
							<div class="eventitem">
								<span class="uv-event-name">Past Event</span>
								<a class="venueurl">AYU Dayclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr, pastDateStr),
				`<html><body></body></html>`,
			},
			statusCodes:        []int{http.StatusOK, http.StatusOK},
			expectedEventCount: 1,
			description:        "Should only include future events",
		},
		{
			name: "Malformed HTML with no event containers",
			mockHTMLResponses: []string{
				`
					<html>
						<body>
							<div class="some-other-class">Not an event</div>
						</body>
					</html>
				`,
			},
			statusCodes:        []int{http.StatusOK},
			expectedEventCount: 0,
			description:        "Should handle HTML without proper event containers",
		},
		{
			name: "Missing artist name in HTML",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				`<html><body></body></html>`,
			},
			statusCodes:        []int{http.StatusOK, http.StatusOK},
			expectedEventCount: 1,
			description:        "Should handle missing artist name gracefully",
		},
		{
			name: "Missing club name in HTML",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Artist Name</span>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				`<html><body></body></html>`,
			},
			statusCodes:        []int{http.StatusOK, http.StatusOK},
			expectedEventCount: 1,
			description:        "Should handle missing club name gracefully",
		},
		{
			name: "Invalid URL format (cannot extract date)",
			mockHTMLResponses: []string{
				`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Test Artist</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/invalid"></a>
							</div>
						</body>
					</html>
				`,
				`<html><body></body></html>`,
			},
			statusCodes:        []int{http.StatusOK, http.StatusOK},
			expectedEventCount: 0,
			description:        "Should handle URLs with invalid date format",
		},
		{
			name: "Malformed date in URL",
			mockHTMLResponses: []string{
				`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Test Artist</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/99999999"></a>
							</div>
						</body>
					</html>
				`,
				`<html><body></body></html>`,
			},
			statusCodes:        []int{http.StatusOK, http.StatusOK},
			expectedEventCount: 0,
			description:        "Should handle malformed dates gracefully",
		},
		{
			name: "Pagination continues across multiple months until empty",
			mockHTMLResponses: []string{
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Event 1</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Event 2</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				fmt.Sprintf(`
					<html>
						<body>
							<div class="eventitem">
								<span class="uv-event-name">Event 3</span>
								<a class="venueurl">Zouk Nightclub</a>
								<a class="uv-boxitem noloader" href="https://zoukgrouplv.com/events/%s"></a>
							</div>
						</body>
					</html>
				`, futureDateStr),
				`<html><body></body></html>`, // Empty stops pagination
			},
			statusCodes:        []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK},
			expectedEventCount: 3,
			description:        "Pagination should continue until empty response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			// Create mock server
			responseIndex := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if responseIndex < len(tt.mockHTMLResponses) {
					statusCode := http.StatusOK
					if responseIndex < len(tt.statusCodes) {
						statusCode = tt.statusCodes[responseIndex]
					}
					w.Header().Set("Content-Type", "text/html")
					w.WriteHeader(statusCode)
					w.Write([]byte(tt.mockHTMLResponses[responseIndex]))
					responseIndex++
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			// Call the function directly with test server URL
			events := scrapeZoukEdmEvents(server.URL + "?date=")

			// Verify event count
			if len(events) != tt.expectedEventCount {
				t.Errorf("%s: Expected %d events, got %d", tt.description, tt.expectedEventCount, len(events))
			}
		})
	}
}
