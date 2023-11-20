package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"net/http"
	"strings"
	"time"
)

/*
Update: 07/24/2023
Zouk added lazy loading to their event page, Sadly I can't hit the events url to getGUID all the event since it only shows
a subset of the events, There are two alternatives, either getGUID puppeter or something similar to literally scroll down to load allvevents then scrape them
OR
fetch the same url that the lazy loaded feature on the event page does, I Implemented this method here, I figured it would be easier then installing a whole new package.
----

I am fetching the url for each month possible, we start at the current month then we hit each month until we don't getGUID any more eventData. I deprecated the older Zouk script,
might come in useful if they decide to remove the lazy loading feature then this won't be needed
*/

func (app *application) fetchZoukEdmEvents(w http.ResponseWriter, r *http.Request) {

	edmEvents := scrapeZoukEdmEvents()
	err := app.writeJSON(w, http.StatusOK, edmEvents, nil)

	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

}

func scrapeZoukEdmEvents() []EdmEvent {
	currentTime := time.Now()
	monthNumber := int(currentTime.Month())
	year := currentTime.Year()
	edmEvents := []EdmEvent{}
	var scrapeurl string
	hasEventItems := true

	c := colly.NewCollector()

	c.OnHTML("body", func(h *colly.HTMLElement) {
		selection := h.DOM

		if selection.Find("div.eventitem").Length() == 0 {
			hasEventItems = false
		}

	})

	c.OnHTML("div.eventitem ", func(h *colly.HTMLElement) {
		selection := h.DOM
		edmEvent := EdmEvent{}
		artistName := selection.Find("span.uv-event-name").Text()
		clubName := selection.Find("a.venueurl").Text()
		edmEvent.Id = getGUID()
		edmEvent.ArtistName = strings.ToLower(artistName)
		edmEvent.ClubName = strings.ToLower(clubName)
		venueTicketurl, _ := selection.Find(".uv-boxitem.noloader").Attr("href")
		edmEvent.TicketUrl = venueTicketurl
		formattedDate, err := formatDateFrom_YYYYMMDD_toRFC3339(extractEventDate(venueTicketurl))

		if err != nil {
			fmt.Println("Error while parsing the date:", err)
		}

		edmEvent.EventDate = formattedDate

		edmEvents = append(edmEvents, edmEvent)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Printf("Error while scraping: %s\n", e.Error())
	})

	c.OnScraped(func(r *colly.Response) {

	})

	for hasEventItems {
		scrapeurl = fmt.Sprintf("https://zoukgrouplv.com/wp-admin/admin-ajax.php?action=uvwp_loadmoreevents&venuegroup=all&caldate=%v-%02d-01", year, monthNumber)
		year, monthNumber = incrementYearMonth(year, monthNumber)
		c.Visit(scrapeurl)
	}
	return edmEvents
}

func incrementYearMonth(year int, monthNumber int) (int, int) {
	if monthNumber >= 12 {
		year++
		monthNumber = 1
	} else {
		monthNumber++
	}
	return year, monthNumber
}
