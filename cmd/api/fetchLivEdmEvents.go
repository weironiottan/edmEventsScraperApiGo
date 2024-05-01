package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"net/http"
	"strings"
)

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
	scrapeurl := "https://www.livnightclub.com/las-vegas/events/?venue=livlasvegas"
	c := colly.NewCollector()
	c.Wait()

	c.OnHTML("div.uv-carousel-lat ", func(h *colly.HTMLElement) {
		selection := h.DOM
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

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Printf("Error while scraping: %s\n", e.Error())
	})

	c.Visit(scrapeurl)

	fmt.Println("Scraping Completed for Liv")
	return edmEvents
}
