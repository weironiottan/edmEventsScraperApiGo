package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"net/http"
	"strings"
)

// Declare a handler which writes a plain-text response with information about the
// application status, operating environment and version.
func (app *application) fetchWynnEdmEvents(w http.ResponseWriter, r *http.Request) {
	edmEvents := scrapeWynnForEdmEvents()
	err := app.writeJSON(w, http.StatusOK, edmEvents, nil)

	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

}

func scrapeWynnForEdmEvents() []EdmEvent {
	edmEvents := []EdmEvent{}
	scrapeurl := "https://www.wynnsocial.com/events/"
	c := colly.NewCollector()
	c.Wait()

	c.OnHTML("div.eventitem ", func(h *colly.HTMLElement) {
		selection := h.DOM
		edmEvent := EdmEvent{}
		artistName := selection.Find("span.uv-events-name").Text()
		clubName := selection.Find("span.venueurl").Text()
		edmEvent.Id = getGUID()
		edmEvent.ArtistName = strings.ToLower(artistName)
		edmEvent.ClubName = strings.ToLower(clubName)
		venueTicketurl, _ := selection.Find("a.uv-btn").Attr("href")
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
		fmt.Println(len(edmEvents), "NOT filtered events")
		edmEvents = filterUnwantedEvents(edmEvents, []string{"wynn field club", "festival", "art of the wild"})
		fmt.Println(len(edmEvents), "filtered events")

	})

	c.Visit(scrapeurl)

	fmt.Println("Scraping Completed for Wynn")
	return edmEvents
}
