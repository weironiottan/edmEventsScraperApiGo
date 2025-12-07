package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

/*
Update: 07/13/2025

Liv has updated their site to use pagination, it's easier now to just call the API directly
with the current date, and then find out what the next pagination date is from the response
it also returns the HTML in the response, so I threw out Colly and just use
goquery to query the HTML
----
*/

func scrapeLivForEdmEvents(url string) []EdmEvent {
	edmEvents := []EdmEvent{}
	currentDate := time.Now().Format("2006-01-02")

	for {
		url := formatPaginatedDateURLWynn(url, currentDate)

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

	fmt.Println(len(edmEvents), "LIV edmEvents")
	fmt.Println("Scraping Completed for LIV!!!")
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
			return
		}

		edmEvent.EventDate = formattedDate

		isPastDate, err := isPastDate(formattedDate)
		if err != nil {
			fmt.Println("Error while parsing the date:", err)
			return
		}

		if !isPastDate {
			edmEvents = append(edmEvents, edmEvent)
		}
	})

	return edmEvents
}

func formatPaginatedDateURLWynn(scrappingUrl string, date string) string {
	formattedURL := fmt.Sprintf(scrappingUrl+"%s&venue=livlasvegas", date)
	return formattedURL
}

type LivEdmEventsResponse struct {
	Agenda       string `json:"agenda"`
	Calendar     string `json:"calendar"`
	List         string `json:"list"`
	Todate       string `json:"todate"`
	Nevents      int    `json:"nevents"`
	Nextloaddate string `json:"nextloaddate"`
}
