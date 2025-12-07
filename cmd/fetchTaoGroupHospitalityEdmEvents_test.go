package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestScrapeTaoGroupHospitalityEdmEvents_Positive tests successful scraping scenarios
func TestScrapeTaoGroupHospitalityEdmEvents_Positive(t *testing.T) {
	// Get future date for testing (30 days from now)
	futureDate := time.Now().AddDate(0, 0, 30)
	futureDateStr := futureDate.Format("01/02/2006")

	tests := []struct {
		name                string
		mockAPIResponses    []mockResponse
		expectedEventCount  int
		expectedFirstArtist string
		expectedFirstClub   string
		validateAllEvents   func(t *testing.T, events []EdmEvent)
	}{
		{
			name: "Single page with one valid event",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[{
						"id": 1,
						"link": "https://taogroup.com/event/tiesto",
						"acf": {
							"event_title": {"display_title": "Tiësto"},
							"event_start_date": "%s 10:00 PM",
							"event_venue": [{"post_title": "Hakkasan - Las Vegas"}]
						}
					}]`, futureDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``}, // End pagination
			},
			expectedEventCount:  1,
			expectedFirstArtist: "tiësto",
			expectedFirstClub:   "hakkasan",
			validateAllEvents: func(t *testing.T, events []EdmEvent) {
				if len(events) != 1 {
					return
				}
				if events[0].TicketUrl != "https://taogroup.com/event/tiesto" {
					t.Errorf("Expected ticket URL 'https://taogroup.com/event/tiesto', got '%s'", events[0].TicketUrl)
				}
				if events[0].Id == "" {
					t.Error("Expected event ID to be generated")
				}
			},
		},
		{
			name: "Single page with multiple events",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[
						{
							"id": 1,
							"link": "https://taogroup.com/event/calvin-harris",
							"acf": {
								"event_title": {"display_title": "Calvin Harris"},
								"event_start_date": "%s 11:00 PM",
								"event_venue": [{"post_title": "Omnia - Las Vegas"}]
							}
						},
						{
							"id": 2,
							"link": "https://taogroup.com/event/david-guetta",
							"acf": {
								"event_title": {"display_title": "David Guetta"},
								"event_start_date": "%s 10:30 PM",
								"event_venue": [{"post_title": "Hakkasan - Las Vegas"}]
							}
						}
					]`, futureDateStr, futureDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount:  2,
			expectedFirstArtist: "calvin harris",
			expectedFirstClub:   "omnia",
		},
		{
			name: "Multiple pages with pagination",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[{
						"id": 1,
						"link": "https://taogroup.com/event/zedd",
						"acf": {
							"event_title": {"display_title": "Zedd"},
							"event_start_date": "%s 10:00 PM",
							"event_venue": [{"post_title": "TAO Nightclub - Las Vegas"}]
						}
					}]`, futureDateStr),
				},
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[{
						"id": 2,
						"link": "https://taogroup.com/event/marshmello",
						"acf": {
							"event_title": {"display_title": "Marshmello"},
							"event_start_date": "%s 11:00 PM",
							"event_venue": [{"post_title": "Marquee - Las Vegas"}]
						}
					}]`, futureDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount:  2,
			expectedFirstArtist: "zedd",
			expectedFirstClub:   "tao nightclub",
		},
		{
			name: "Artist names are lowercased",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[{
						"id": 1,
						"link": "https://taogroup.com/event/steve-aoki",
						"acf": {
							"event_title": {"display_title": "STEVE AOKI"},
							"event_start_date": "%s 09:00 PM",
							"event_venue": [{"post_title": "Jewel - Las Vegas"}]
						}
					}]`, futureDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount:  1,
			expectedFirstArtist: "steve aoki",
			expectedFirstClub:   "jewel",
		},
		{
			name: "Las Vegas suffix is removed from venue names",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[{
						"id": 1,
						"link": "https://taogroup.com/event/test",
						"acf": {
							"event_title": {"display_title": "Alesso"},
							"event_start_date": "%s 10:00 PM",
							"event_venue": [{"post_title": "XS Nightclub - Las Vegas"}]
						}
					}]`, futureDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount:  1,
			expectedFirstArtist: "alesso",
			expectedFirstClub:   "xs nightclub",
		},
		{
			name: "LAVO Italian Restaurant events are filtered (case insensitive)",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[{
						"id": 1,
						"link": "https://taogroup.com/event/restaurant-event",
						"acf": {
							"event_title": {"display_title": "Dinner Show"},
							"event_start_date": "%s 07:00 PM",
							"event_venue": [{"post_title": "Lavo Italian Restaurant"}]
						}
					}]`, futureDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount: 0,
		},
		{
			name: "Empty response returns no events",
			mockAPIResponses: []mockResponse{
				{statusCode: http.StatusOK, body: `[]`},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount: 0,
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
			events := scrapeTaoGroupHospitalityEdmEvents(server.URL + "?")

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

// TestScrapeTaoGroupHospitalityEdmEvents_Negative tests error scenarios and edge cases
func TestScrapeTaoGroupHospitalityEdmEvents_Negative(t *testing.T) {
	// Get past date for testing (30 days ago)
	pastDate := time.Now().AddDate(0, 0, -30)
	pastDateStr := pastDate.Format("01/02/2006")

	// Get future date
	futureDate := time.Now().AddDate(0, 0, 30)
	futureDateStr := futureDate.Format("01/02/2006")

	tests := []struct {
		name               string
		mockAPIResponses   []mockResponse
		expectedEventCount int
		description        string
	}{
		{
			name: "Server returns 404 on first page",
			mockAPIResponses: []mockResponse{
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount: 0,
			description:        "Should handle server 404 gracefully",
		},
		{
			name: "Server returns 500 error",
			mockAPIResponses: []mockResponse{
				{statusCode: http.StatusInternalServerError, body: ``},
			},
			expectedEventCount: 0,
			description:        "Should handle server errors gracefully",
		},
		{
			name: "Invalid JSON response",
			mockAPIResponses: []mockResponse{
				{statusCode: http.StatusOK, body: `{invalid json}`},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount: 0,
			description:        "Should handle malformed JSON without crashing",
		},
		{
			name: "Past events are filtered out",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[{
						"id": 1,
						"link": "https://taogroup.com/event/past-event",
						"acf": {
							"event_title": {"display_title": "Past Event"},
							"event_start_date": "%s 10:00 PM",
							"event_venue": [{"post_title": "Hakkasan - Las Vegas"}]
						}
					}]`, pastDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount: 0,
			description:        "Events with past dates should be filtered",
		},
		{
			name: "LAVO restaurant events are filtered out",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[{
						"id": 1,
						"link": "https://taogroup.com/event/restaurant",
						"acf": {
							"event_title": {"display_title": "Dinner Event"},
							"event_start_date": "%s 07:00 PM",
							"event_venue": [{"post_title": "LAVO Italian Restaurant Las Vegas"}]
						}
					}]`, futureDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount: 0,
			description:        "LAVO restaurant events should be filtered",
		},
		{
			name: "Mix of valid and filtered events",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[
						{
							"id": 1,
							"link": "https://taogroup.com/event/valid",
							"acf": {
								"event_title": {"display_title": "Valid Event"},
								"event_start_date": "%s 10:00 PM",
								"event_venue": [{"post_title": "Hakkasan - Las Vegas"}]
							}
						},
						{
							"id": 2,
							"link": "https://taogroup.com/event/past",
							"acf": {
								"event_title": {"display_title": "Past Event"},
								"event_start_date": "%s 10:00 PM",
								"event_venue": [{"post_title": "Omnia - Las Vegas"}]
							}
						},
						{
							"id": 3,
							"link": "https://taogroup.com/event/restaurant",
							"acf": {
								"event_title": {"display_title": "Restaurant Event"},
								"event_start_date": "%s 07:00 PM",
								"event_venue": [{"post_title": "LAVO Italian Restaurant Las Vegas"}]
							}
						}
					]`, futureDateStr, pastDateStr, futureDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount: 1,
			description:        "Should only include valid future nightclub events",
		},
		{
			name: "Missing event venue",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: fmt.Sprintf(`[{
						"id": 1,
						"link": "https://taogroup.com/event/test",
						"acf": {
							"event_title": {"display_title": "Test Event"},
							"event_start_date": "%s 10:00 PM",
							"event_venue": []
						}
					}]`, futureDateStr),
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount: 0,
			description:        "Events with missing venue should be skipped",
		},
		{
			name: "Malformed date string",
			mockAPIResponses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body: `[{
						"id": 1,
						"link": "https://taogroup.com/event/test",
						"acf": {
							"event_title": {"display_title": "Test Event"},
							"event_start_date": "invalid-date",
							"event_venue": [{"post_title": "Hakkasan - Las Vegas"}]
						}
					}]`,
				},
				{statusCode: http.StatusNotFound, body: ``},
			},
			expectedEventCount: 0,
			description:        "Events with invalid dates should be handled gracefully",
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
			events := scrapeTaoGroupHospitalityEdmEvents(server.URL + "?")

			// Verify event count
			if len(events) != tt.expectedEventCount {
				t.Errorf("%s: Expected %d events, got %d", tt.description, tt.expectedEventCount, len(events))
			}
		})
	}
}

// Helper types

type mockResponse struct {
	statusCode int
	body       string
}
