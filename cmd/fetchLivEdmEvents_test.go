package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestScrapeLivForEdmEvents_Positive tests successful scraping scenarios
func TestScrapeLivForEdmEvents_Positive(t *testing.T) {
	// Get future date for testing (30 days from now)
	futureDate := time.Now().AddDate(0, 0, 30)
	futureDateStr := futureDate.Format("20060102") // YYYYMMDD format for URL extraction

	tests := []struct {
		name                string
		mockAPIResponses    []mockLivResponse
		expectedEventCount  int
		expectedFirstArtist string
		expectedFirstClub   string
		validateAllEvents   func(t *testing.T, events []EdmEvent)
	}{
		{
			name: "Single page with one valid event",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Tiësto</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`, futureDateStr),
				},
			},
			expectedEventCount:  1,
			expectedFirstArtist: "tiësto",
			expectedFirstClub:   "liv nightclub",
			validateAllEvents: func(t *testing.T, events []EdmEvent) {
				if len(events) != 1 {
					return
				}
				if events[0].TicketUrl != fmt.Sprintf("https://livnightclub.com/events/%s", futureDateStr) {
					t.Errorf("Expected ticket URL 'https://livnightclub.com/events/%s', got '%s'", futureDateStr, events[0].TicketUrl)
				}
				if events[0].Id == "" {
					t.Error("Expected event ID to be generated")
				}
			},
		},
		{
			name: "Single page with multiple events",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Calvin Harris</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div><div class='uv-carousel-lat'><h3 class='uv-event-name-title'>David Guetta</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 2,
						"nextloaddate": ""
					}`, futureDateStr, futureDateStr),
				},
			},
			expectedEventCount:  2,
			expectedFirstArtist: "calvin harris",
			expectedFirstClub:   "liv nightclub",
		},
		{
			name: "Multiple pages with pagination",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Zedd</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": "2025-12-15"
					}`, futureDateStr),
				},
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Marshmello</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`, futureDateStr),
				},
			},
			expectedEventCount:  2,
			expectedFirstArtist: "zedd",
			expectedFirstClub:   "liv nightclub",
		},
		{
			name: "Artist names are lowercased",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>STEVE AOKI</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`, futureDateStr),
				},
			},
			expectedEventCount:  1,
			expectedFirstArtist: "steve aoki",
			expectedFirstClub:   "liv nightclub",
		},
		{
			name: "Club names are lowercased",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Alesso</h3><div class='uwsvenuename'>LIV NIGHTCLUB</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`, futureDateStr),
				},
			},
			expectedEventCount:  1,
			expectedFirstArtist: "alesso",
			expectedFirstClub:   "liv nightclub",
		},
		{
			name: "Empty agenda returns no events",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: `{
						"agenda": "",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 0,
						"nextloaddate": ""
					}`,
				},
			},
			expectedEventCount: 0,
		},
		{
			name: "Pagination stops when nevents is 0",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>First Artist</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": "2025-12-15"
					}`, futureDateStr),
				},
				{
					statusCode: http.StatusOK,
					body: `{
						"agenda": "",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 0,
						"nextloaddate": ""
					}`,
				},
			},
			expectedEventCount:  1,
			expectedFirstArtist: "first artist",
			expectedFirstClub:   "liv nightclub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			responseIndex := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if responseIndex < len(tt.mockAPIResponses) {
					mockResp := tt.mockAPIResponses[responseIndex]
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(mockResp.statusCode)
					w.Write([]byte(mockResp.body))
					responseIndex++
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			// Call the function directly with test server URL
			events := scrapeLivForEdmEvents(server.URL + "?date=")

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

// TestScrapeLivForEdmEvents_Negative tests error scenarios and edge cases
func TestScrapeLivForEdmEvents_Negative(t *testing.T) {
	// Get past date for testing (30 days ago)
	pastDate := time.Now().AddDate(0, 0, -30)
	pastDateStr := pastDate.Format("20060102") // YYYYMMDD format for URL extraction

	// Get future date
	futureDate := time.Now().AddDate(0, 0, 30)
	futureDateStr := futureDate.Format("20060102")

	tests := []struct {
		name               string
		mockAPIResponses   []mockLivResponse
		expectedEventCount int
		description        string
	}{
		{
			name: "Server returns error on first request",
			mockAPIResponses: []mockLivResponse{
				{statusCode: http.StatusInternalServerError, body: ``},
			},
			expectedEventCount: 0,
			description:        "Should handle server errors gracefully",
		},
		{
			name: "Invalid JSON response",
			mockAPIResponses: []mockLivResponse{
				{statusCode: http.StatusOK, body: `{invalid json}`},
			},
			expectedEventCount: 0,
			description:        "Should handle malformed JSON without crashing",
		},
		{
			name: "Past events are filtered out",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Past Event</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`, pastDateStr),
				},
			},
			expectedEventCount: 0,
			description:        "Events with past dates should be filtered",
		},
		{
			name: "Mix of valid and past events",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Valid Event</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div><div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Past Event</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 2,
						"nextloaddate": ""
					}`, futureDateStr, pastDateStr),
				},
			},
			expectedEventCount: 1,
			description:        "Should only include future events",
		},
		{
			name: "Malformed HTML with no event containers",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: `{
						"agenda": "<div class='some-other-class'>Not an event</div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`,
				},
			},
			expectedEventCount: 0,
			description:        "Should handle HTML without proper event containers",
		},
		{
			name: "Missing artist name in HTML",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`, futureDateStr),
				},
			},
			expectedEventCount: 1,
			description:        "Should handle missing artist name gracefully",
		},
		{
			name: "Missing club name in HTML",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Artist Name</h3><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`, futureDateStr),
				},
			},
			expectedEventCount: 1,
			description:        "Should handle missing club name gracefully",
		},
		{
			name: "Invalid URL format (cannot extract date)",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: `{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Test Artist</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/invalid'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`,
				},
			},
			expectedEventCount: 0,
			description:        "Should handle URLs with invalid date format",
		},
		{
			name: "Malformed date in URL",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: `{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Test Artist</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/99999999'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`,
				},
			},
			expectedEventCount: 0,
			description:        "Should handle malformed dates gracefully",
		},
		{
			name: "Empty nextloaddate stops pagination",
			mockAPIResponses: []mockLivResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`{
						"agenda": "<div class='uv-carousel-lat'><h3 class='uv-event-name-title'>Artist</h3><div class='uwsvenuename'>LIV Nightclub</div><a class='hd-link' href='https://livnightclub.com/events/%s'></a></div>",
						"calendar": "",
						"list": "",
						"todate": "",
						"nevents": 1,
						"nextloaddate": ""
					}`, futureDateStr),
				},
			},
			expectedEventCount: 1,
			description:        "Pagination should stop when nextloaddate is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			// Create mock server
			responseIndex := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if responseIndex < len(tt.mockAPIResponses) {
					mockResp := tt.mockAPIResponses[responseIndex]
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(mockResp.statusCode)
					w.Write([]byte(mockResp.body))
					responseIndex++
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			// Call the function directly with test server URL
			events := scrapeLivForEdmEvents(server.URL + "?date=")

			// Verify event count
			if len(events) != tt.expectedEventCount {
				t.Errorf("%s: Expected %d events, got %d", tt.description, tt.expectedEventCount, len(events))
			}
		})
	}
}

// Helper types

type mockLivResponse struct {
	statusCode int
	body       string
}
