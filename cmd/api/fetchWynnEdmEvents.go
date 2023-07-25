package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"net/http"
	"regexp"
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

	c.OnHTML("div.eventitem ", func(h *colly.HTMLElement) {
		selection := h.DOM
		edmEvent := EdmEvent{}
		artistName := selection.Find("span.uv-events-name").Text()
		clubName := selection.Find("a.venueurl").Text()
		edmEvent.ArtistName = artistName
		edmEvent.ClubName = clubName
		venueTicketurl, _ := selection.Find("a.uv-btn").Attr("href")
		edmEvent.TicketUrl = venueTicketurl
		edmEvent.EventDate = extractDate(venueTicketurl)

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
		edmEvents = filterUnwantedEvents(edmEvents)
		fmt.Println(len(edmEvents), "filtered events")

	})

	c.Visit(scrapeurl)
	return edmEvents
}

func extractDate(url string) string {
	regexPattern := regexp.MustCompile(`\d+`)
	extractedData := regexPattern.FindStringSubmatch(url)
	extractedDigits := extractedData[0]
	extractedDate := extractedDigits[len(extractedDigits)-8:]
	return extractedDate
}

func filterUnwantedEvents(edmEvents []EdmEvent) []EdmEvent {
	unWantedEvents := []string{"wynn field club", "festival", "art of the wild"}
	var filteredEdmEvents []EdmEvent

	for _, edmEvent := range edmEvents {
		isWantedClubName := filterEvent(edmEvent.ClubName, unWantedEvents)
		isWantedArtistName := filterEvent(edmEvent.ArtistName, unWantedEvents)
		if isWantedClubName && isWantedArtistName {
			filteredEdmEvents = append(filteredEdmEvents, edmEvent)
		}
	}
	return filteredEdmEvents
}

func filterEvent(str string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(strings.ToLower(str), substring) {
			return false
		}
	}
	return true
}
