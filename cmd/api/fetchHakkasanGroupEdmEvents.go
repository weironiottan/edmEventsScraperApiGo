package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (app *application) fetchHakkasanGroupEdmEvents(w http.ResponseWriter, r *http.Request) {

	edmEvents := scrapeHakkasanGroupEdmEvents()
	err := app.writeJSON(w, http.StatusOK, edmEvents, nil)

	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

}

func scrapeHakkasanGroupEdmEvents() []EdmEvent {
	var hakassanGroupEdmEvents HakassanGroupEdmEvents
	edmEvents := []EdmEvent{}

	client := &http.Client{}

	// Create a GET request
	request, err := http.NewRequest("GET", "https://data.portaldriver.engineering/events.json", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
	}

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
	}
	jsonData := extractJSONPData(string(body))

	err = json.Unmarshal([]byte(jsonData), &hakassanGroupEdmEvents)
	if err != nil {
		fmt.Println("Error:", err)
	}

	location := "las vegas"

	for _, hakassanEvent := range hakassanGroupEdmEvents.Data {
		if strings.Contains(strings.ToLower(hakassanEvent.Location), location) {
			edmEvent := EdmEvent{}
			edmEvent.ArtistName = hakassanEvent.VenueTitle
			edmEvent.ClubName = hakassanEvent.Title
			edmEvent.EventDate = hakassanEvent.Date
			edmEvent.TicketUrl = fmt.Sprintf("https://events.taogroup.com/events/%v", hakassanEvent.Id)
			edmEvents = append(edmEvents, edmEvent)
		}

	}
	return edmEvents
}

func extractJSONPData(jsonp string) string {
	startIndex := strings.Index(jsonp, "(") + 1
	endIndex := strings.LastIndex(jsonp, ")")
	if startIndex >= endIndex {
		return ""
	}
	return jsonp[startIndex:endIndex]
}

type HakassanGroupEdmEvents struct {
	Ref  string `json:"ref"`
	Data []struct {
		Id          int    `json:"id"`
		Title       string `json:"title"`
		Location    string `json:"location"`
		Description string `json:"description"`
		TimeZone    string `json:"time_zone"`
		VenueId     int    `json:"venue_id"`
		VenueTitle  string `json:"venue_title"`
		Region      struct {
			Id        int    `json:"id"`
			Name      string `json:"name"`
			Slug      string `json:"slug"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		} `json:"region"`
		Date               string      `json:"date"`
		Open               time.Time   `json:"open"`
		Close              time.Time   `json:"close"`
		ClosedMessage      *string     `json:"closed_message"`
		Active             *bool       `json:"active"`
		ShowInCalendars    bool        `json:"show_in_calendars"`
		PublicReservations *bool       `json:"public_reservations"`
		PublicGuestlists   *bool       `json:"public_guestlists"`
		HasPublicTickets   bool        `json:"has_public_tickets"`
		VIPURL             *string     `json:"VIP_URL"`
		DayOfTheWeek       int         `json:"dayOfTheWeek"`
		TagList            []string    `json:"tag_list"`
		TicketsURL         interface{} `json:"tickets_URL"`
		FlyerUrl           string      `json:"flyer_url"`
		ArtistEvent        []struct {
			Id         int    `json:"id"`
			FriendlyId string `json:"friendly_id"`
			Name       string `json:"name"`
			Area       string `json:"area"`
		} `json:"artist_event"`
		Headliner *int `json:"headliner"`
	} `json:"data"`
	DataType string `json:"dataType"`
}
