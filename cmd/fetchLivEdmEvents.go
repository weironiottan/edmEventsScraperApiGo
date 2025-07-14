package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"time"
)

/*
Update: 07/13/2025

Liv has updated their site to use pagination, it's easier now to just call the API directly
with the current date, and then find out what the next pagination date is from the response
it also returns the HTML in the response, so I threw out Colly and just use
goquery to query the HTML
----
*/
func (app *application) fetchLivEdmEvents(w http.ResponseWriter, r *http.Request) {
	edmEvents := scrapeLivForEdmEvents()
	err := app.writeJSON(w, http.StatusOK, edmEvents, nil)

	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

}

func scrapeLivForEdmEvents() []EdmEvent {
	edmEvents := []EdmEvent{}
	currentDate := time.Now().Format("2006-01-02")

	for {
		url := fmt.Sprintf("https://www.livnightclub.com/wp-admin/admin-ajax.php?action=uvpx&uvaction=uwspx_loadevents&date=%s&venue=livlasvegas", currentDate)

		resp, err := http.Get(url)
		if err != nil {
			break
		}

		// JSON unmarshal
		var livEdmEventsResponse LivEdmEventsResponse
		json.NewDecoder(resp.Body).Decode(&livEdmEventsResponse)
		resp.Body.Close()

		// GoQuery on the HTML string
		events := parseHTMLWithGoQuery(livEdmEventsResponse.Agenda)
		edmEvents = append(edmEvents, events...)

		// Pagination logic
		if livEdmEventsResponse.Nextloaddate == "" || livEdmEventsResponse.Nevents < 1 {
			break
		}
		currentDate = livEdmEventsResponse.Nextloaddate
	}
	return edmEvents
}

func parseHTMLWithGoQuery(htmlContent string) []EdmEvent {
	var edmEvents []EdmEvent

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return edmEvents
	}

	doc.Find("div.uv-carousel-lat").Each(func(i int, selection *goquery.Selection) {
		edmEvent := EdmEvent{}
		artistName := selection.Find("h3.uv-event-name-title").Text()
		clubName := selection.Find("div.uwsvenuename").Text()
		edmEvent.Id = getGUID()
		edmEvent.ArtistName = strings.ToLower(artistName)
		edmEvent.ClubName = strings.ToLower(clubName)
		venueTicketurl, _ := selection.Find("a.hd-link").Attr("href")
		edmEvent.TicketUrl = venueTicketurl
		formattedDate, err := formatDateFrom_YYYYMMDD_toRFC3339(extractEventDate(venueTicketurl))

		if err != nil {
			fmt.Println("Error while parsing the date:", err)
		}

		edmEvent.EventDate = formattedDate

		isPastDate, err := isPastDate(formattedDate)
		if err != nil {
			fmt.Println("Error while parsing the date:", err)
		}

		if !isPastDate {
			edmEvents = append(edmEvents, edmEvent)
		}
	})

	return edmEvents
}

type LivEdmEventsResponse struct {
	Agenda       string `json:"agenda"`
	Calendar     string `json:"calendar"`
	List         string `json:"list"`
	Todate       string `json:"todate"`
	Nevents      int    `json:"nevents"`
	Nextloaddate string `json:"nextloaddate"`
}
